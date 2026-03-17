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

	// According to user research, the trigger price comes perfectly formatted from Helper if defined as string,
	// Since in Go we still pass numbers, we standardize the transmission format as strings avoiding sci notation.
	triggerPriceStr := formatPrice(req.TriggerPrice)
	params.Add("triggerPrice", triggerPriceStr)
	
	params.Add("algoType", "CONDITIONAL")

	if req.ClosePosition {
		params.Add("closePosition", "true")
	} else {
		qtyStr := formatQuantity(req.Quantity)
		params.Add("quantity", qtyStr)
	}

	// Add timestamp in milliseconds
	params.Add("timestamp", fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)))

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

// generateSignature generates HMAC-SHA256 signature for Binance API requests
func generateSignature(secretKey string, params url.Values) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(params.Encode()))
	return hex.EncodeToString(mac.Sum(nil))
}
