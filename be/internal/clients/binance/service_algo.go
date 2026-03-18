package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// PlaceAlgoOrder places a conditional algo order via direct HTTP request (useful for ClosePosition=true Stop/Take Profit)
// This implements placing TP/SL directly because go-binance library lacks complete native support for the new `/fapi/v1/algoOrder` spec.
func (c *Client) PlaceAlgoOrder(ctx context.Context, req *PlaceAlgoOrderRequest) (*AlgoOrderResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	endpoint := "/fapi/v1/algoOrder"

	// Formatting payload params
	params := url.Values{}
	params.Add("symbol", req.Symbol)
	params.Add("side", string(req.Side))
	params.Add("type", string(req.Type))

	// Convert trigger price to string without changing decimal places
	params.Add("triggerPrice", formatPrice(req.TriggerPrice))

	params.Add("algoType", "CONDITIONAL")

	if req.ClosePosition {
		params.Add("closePosition", "true")
	} else {
		// Convert quantity to string without changing decimal places
		params.Add("quantity", formatQuantity(req.Quantity))
	}

	// Add timestamp and recvWindow in milliseconds
	params.Add("recvWindow", fmt.Sprintf("%d", c.config.RecvWindow))
	params.Add("timestamp", fmt.Sprintf("%d", c.getTimestamp()))

	// Sign payload via HMAC-SHA256
	signature := generateSignature(c.config.SecretKey, params)
	params.Add("signature", signature)

	// Combine URL
	reqURL := fmt.Sprintf("%s%s?%s", c.config.BaseURL, endpoint, params.Encode())

	// Prepare HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create algo order request: %v", ErrOrderFailed, err)
	}

	// Set required API Key header
	httpReq.Header.Set("X-MBX-APIKEY", c.config.APIKey)

	httpClient := &http.Client{Timeout: c.config.Timeout}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to send algo order request: %v", ErrOrderFailed, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read algo order response: %v", ErrOrderFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrOrderFailed, resp.StatusCode, string(body))
	}

	var parsedResp AlgoOrderResponse
	if err := json.Unmarshal(body, &parsedResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse algo order response: %v", ErrOrderFailed, err)
	}

	// On success `/fapi/v1/algoOrder` returns AlgoID directly and no `code`.
	// On error, it returns `code` and `msg`.
	if parsedResp.Code != 0 && parsedResp.AlgoID == 0 {
		return nil, fmt.Errorf("%w: API error code %d - %s", ErrOrderFailed, parsedResp.Code, parsedResp.Msg)
	}

	return &parsedResp, nil
}

// GetAlgoOrder fetches a specific algo order by ID using /fapi/v1/allAlgoOrders
func (c *Client) GetAlgoOrder(ctx context.Context, req *GetAlgoOrdersRequest) (*AlgoOrderResponse, error) {
	endpoint := "/fapi/v1/allAlgoOrders"

	params := url.Values{}
	params.Add("symbol", req.Symbol)
	if req.AlgoID > 0 {
		params.Add("algoId", fmt.Sprintf("%d", req.AlgoID))
	}

	params.Add("recvWindow", fmt.Sprintf("%d", c.config.RecvWindow))
	params.Add("timestamp", fmt.Sprintf("%d", c.getTimestamp()))
	signature := generateSignature(c.config.SecretKey, params)
	params.Add("signature", signature)

	reqURL := fmt.Sprintf("%s%s?%s", c.config.BaseURL, endpoint, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create get algo order request: %v", ErrOrderFailed, err)
	}

	httpReq.Header.Set("X-MBX-APIKEY", c.config.APIKey)

	httpClient := &http.Client{Timeout: c.config.Timeout}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to send get algo order request: %v", ErrOrderFailed, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read get algo order response: %v", ErrOrderFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrOrderFailed, resp.StatusCode, string(body))
	}

	// Check if it returned an error object instead of an array
	var errResp struct {
		Code int64  `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Code != 0 {
		return nil, fmt.Errorf("%w: API error code %d - %s", ErrOrderFailed, errResp.Code, errResp.Msg)
	}

	var parsedResp []AlgoOrderResponse
	if err := json.Unmarshal(body, &parsedResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse get algo order response: %v", ErrOrderFailed, err)
	}

	if len(parsedResp) == 0 {
		return nil, fmt.Errorf("%w: algo order not found", ErrOrderFailed)
	}

	// Return the specific order based on algoId
	for _, order := range parsedResp {
		if order.AlgoID == req.AlgoID {
			return &order, nil
		}
	}

	return &parsedResp[0], nil
}

// GetOpenAlgoOrders fetches all open algo orders for a given symbol
func (c *Client) GetOpenAlgoOrders(ctx context.Context, symbol string) ([]AlgoOrderResponse, error) {
	endpoint := "/fapi/v1/openAlgoOrders"

	params := url.Values{}
	if symbol != "" {
		params.Add("symbol", symbol)
	}

	// Add recvWindow and timestamp
	params.Add("recvWindow", fmt.Sprintf("%d", c.config.RecvWindow))
	params.Add("timestamp", fmt.Sprintf("%d", c.getTimestamp()))

	// Generate signature BEFORE adding signature to params
	signature := generateSignature(c.config.SecretKey, params)
	params.Add("signature", signature)

	reqURL := fmt.Sprintf("%s%s?%s", c.config.BaseURL, endpoint, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create get open algo orders request: %v", ErrOrderFailed, err)
	}

	httpReq.Header.Set("X-MBX-APIKEY", c.config.APIKey)

	httpClient := &http.Client{Timeout: c.config.Timeout}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to send get open algo orders request: %v", ErrOrderFailed, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read get open algo orders response: %v", ErrOrderFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrOrderFailed, resp.StatusCode, string(body))
	}

	// Check if it returned an error object instead of an array
	var errResp struct {
		Code int64  `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Code != 0 {
		return nil, fmt.Errorf("%w: API error code %d - %s", ErrOrderFailed, errResp.Code, errResp.Msg)
	}

	var parsedResp []AlgoOrderResponse
	if err := json.Unmarshal(body, &parsedResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse get open algo orders response: %v", ErrOrderFailed, err)
	}

	return parsedResp, nil
}

// CancelAlgoOrder cancels a conditional algo order via direct HTTP request
func (c *Client) CancelAlgoOrder(ctx context.Context, req *CancelAlgoOrderRequest) (*AlgoOrderResponse, error) {
	endpoint := "/fapi/v1/algoOrder"

	params := url.Values{}
	params.Add("symbol", req.Symbol)
	params.Add("algoId", fmt.Sprintf("%d", req.AlgoID))
	params.Add("recvWindow", fmt.Sprintf("%d", c.config.RecvWindow))
	params.Add("timestamp", fmt.Sprintf("%d", c.getTimestamp()))

	signature := generateSignature(c.config.SecretKey, params)
	params.Add("signature", signature)

	reqURL := fmt.Sprintf("%s%s?%s", c.config.BaseURL, endpoint, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create cancel algo order request: %v", ErrOrderFailed, err)
	}

	httpReq.Header.Set("X-MBX-APIKEY", c.config.APIKey)

	httpClient := &http.Client{Timeout: c.config.Timeout}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to send cancel algo order request: %v", ErrOrderFailed, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read cancel algo order response: %v", ErrOrderFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrOrderFailed, resp.StatusCode, string(body))
	}

	var parsedResp AlgoOrderResponse
	if err := json.Unmarshal(body, &parsedResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse cancel algo order response: %v", ErrOrderFailed, err)
	}

	if parsedResp.Code != 0 && parsedResp.AlgoID == 0 {
		return nil, fmt.Errorf("%w: API error code %d - %s", ErrOrderFailed, parsedResp.Code, parsedResp.Msg)
	}

	return &parsedResp, nil
}

// generateSignature generates HMAC-SHA256 signature for Binance API requests
// Parameters must be sorted alphabetically by key for Binance API compatibility
func generateSignature(secretKey string, params url.Values) string {
	// Build query string with sorted keys (Binance requires alphabetical order)
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sortedParams strings.Builder
	for i, k := range keys {
		if i > 0 {
			sortedParams.WriteByte('&')
		}
		// Use raw key and value (not double-encoded)
		sortedParams.WriteString(k)
		sortedParams.WriteByte('=')
		sortedParams.WriteString(params.Get(k))
	}

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(sortedParams.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

// getTimestamp returns current timestamp in milliseconds with server time offset
func (c *Client) getTimestamp() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) + c.config.ServerTimeOffset
}
