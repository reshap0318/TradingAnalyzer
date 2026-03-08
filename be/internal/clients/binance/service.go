package binance

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// ============================================================================
// Market Data Services
// ============================================================================

// GetPrice gets current price for a symbol
// Cached in Redis for 5 seconds
func (c *Client) GetPrice(symbol string) (*PriceInfo, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("binance:futures:price:%s", symbol)

	// Try cache first
	if c.IsCacheAvailable() {
		var cachedPrice PriceInfo
		err := c.cache.GetJSON(ctx, cacheKey, &cachedPrice)
		if err == nil && cachedPrice.Price > 0 {
			return &cachedPrice, nil
		}
	}

	// Cache miss - fetch from API
	resp, err := c.apiClient.NewPremiumIndexService().Symbol(symbol).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get price for %s: %v", ErrOrderFailed, symbol, err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("%w: no data found for symbol: %s", ErrInvalidSymbol, symbol)
	}

	priceInfo := &PriceInfo{
		Symbol: symbol,
		Price:  parseFloat(resp[0].MarkPrice),
		Time:   resp[0].Time,
	}

	// Cache for 5 seconds
	if c.IsCacheAvailable() {
		_ = c.cache.SetJSON(ctx, cacheKey, priceInfo, 5*time.Second)
	}

	return priceInfo, nil
}

// GetKlines gets kline/candlestick data for a symbol
// Cached in Redis for 30 seconds
func (c *Client) GetKlines(symbol string, interval string, limit int) ([]KlineInfo, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("binance:futures:klines:%s:%s:%d", symbol, interval, limit)

	// Try cache first
	if c.IsCacheAvailable() {
		var cachedKlines []KlineInfo
		err := c.cache.GetJSON(ctx, cacheKey, &cachedKlines)
		if err == nil && len(cachedKlines) > 0 {
			return cachedKlines, nil
		}
	}

	// Cache miss - fetch from API
	resp, err := c.apiClient.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(limit).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: failed to get klines for %s: %v", ErrOrderFailed, symbol, err)
	}

	klines := make([]KlineInfo, len(resp))
	for i, k := range resp {
		klines[i] = KlineInfo{
			OpenTime:  k.OpenTime,
			Open:      parseFloat(k.Open),
			High:      parseFloat(k.High),
			Low:       parseFloat(k.Low),
			Close:     parseFloat(k.Close),
			Volume:    parseFloat(k.Volume),
			CloseTime: k.CloseTime,
		}
	}

	// Cache for 30 seconds
	if c.IsCacheAvailable() {
		_ = c.cache.SetJSON(ctx, cacheKey, klines, 30*time.Second)
	}

	return klines, nil
}

// GetKlinesWithStartTime gets kline/candlestick data starting from a specific time
// Used for historical data fetching (e.g., backtesting)
// No caching - historical data is fetched once and processed in memory
func (c *Client) GetKlinesWithStartTime(symbol string, interval string, limit int, startTimeMs int64) ([]KlineInfo, error) {
	ctx := context.Background()

	resp, err := c.apiClient.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		StartTime(startTimeMs).
		Limit(limit).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: failed to get klines for %s %s: %v", ErrOrderFailed, symbol, interval, err)
	}

	klines := make([]KlineInfo, len(resp))
	for i, k := range resp {
		klines[i] = KlineInfo{
			OpenTime:  k.OpenTime,
			Open:      parseFloat(k.Open),
			High:      parseFloat(k.High),
			Low:       parseFloat(k.Low),
			Close:     parseFloat(k.Close),
			Volume:    parseFloat(k.Volume),
			CloseTime: k.CloseTime,
		}
	}

	return klines, nil
}

// GetAllPrices gets current prices for all symbols
func (c *Client) GetAllPrices() ([]PriceInfo, error) {
	ctx := context.Background()
	resp, err := c.apiClient.NewPremiumIndexService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get all prices: %v", ErrOrderFailed, err)
	}

	prices := make([]PriceInfo, len(resp))
	for i, p := range resp {
		prices[i] = PriceInfo{
			Symbol: p.Symbol,
			Price:  parseFloat(p.MarkPrice),
			Time:   p.Time,
		}
	}

	return prices, nil
}

// ============================================================================
// Multi-Timeframe Market Data Services
// ============================================================================

// GetMultiKlines gets kline data for multiple timeframes in parallel
// Returns map[interval][OHLCV]data
func (c *Client) GetMultiKlines(symbol string, requests []MultiKlineRequest) (map[string][]KlineInfo, error) {
	// Create channels for results and errors
	type klineResult struct {
		interval string
		klines   []KlineInfo
		err      error
	}

	resultChan := make(chan klineResult, len(requests))

	// Launch goroutines for each timeframe
	for _, req := range requests {
		go func(interval string, limit int) {
			klines, err := c.GetKlines(symbol, interval, limit)
			resultChan <- klineResult{
				interval: interval,
				klines:   klines,
				err:      err,
			}
		}(req.Interval, req.Limit)
	}

	// Collect results
	results := make(map[string][]KlineInfo, len(requests))
	var firstErr error

	for i := 0; i < len(requests); i++ {
		result := <-resultChan
		if result.err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("failed to get klines for %s: %v", result.interval, result.err)
			}
			continue
		}
		results[result.interval] = result.klines
	}

	if firstErr != nil && len(results) == 0 {
		return nil, firstErr
	}

	return results, firstErr
}

// GetMultiKlinesOHLCV gets OHLCV data for multiple timeframes in parallel
// Returns map[interval][OHLCV]data where OHLCV = [open, high, low, close, volume]
func (c *Client) GetMultiKlinesOHLCV(symbol string, requests []MultiKlineRequest) (map[string][][]float64, error) {
	// Get klines first
	klinesMap, err := c.GetMultiKlines(symbol, requests)
	if err != nil {
		return nil, err
	}

	// Convert to OHLCV format
	ohlcvMap := make(map[string][][]float64, len(klinesMap))

	for interval, klines := range klinesMap {
		ohlcv := make([][]float64, len(klines))
		for i, k := range klines {
			ohlcv[i] = []float64{
				k.Open,
				k.High,
				k.Low,
				k.Close,
				k.Volume,
			}
		}
		ohlcvMap[interval] = ohlcv
	}

	return ohlcvMap, nil
}

// ============================================================================
// Account Services
// ============================================================================

// GetAccountInfo gets futures account information
func (c *Client) GetAccountInfo() (*AccountInfo, error) {
	ctx := context.Background()
	resp, err := c.apiClient.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get account info: %v", ErrOrderFailed, err)
	}

	balances := make([]BalanceInfo, len(resp.Assets))
	for i, b := range resp.Assets {
		balances[i] = BalanceInfo{
			Asset:                  b.Asset,
			WalletBalance:          parseFloat(b.WalletBalance),
			UnrealizedProfit:       parseFloat(b.UnrealizedProfit),
			MarginBalance:          parseFloat(b.MarginBalance),
			MaintMargin:            parseFloat(b.MaintMargin),
			InitialMargin:          parseFloat(b.InitialMargin),
			PositionInitialMargin:  parseFloat(b.PositionInitialMargin),
			OpenOrderInitialMargin: parseFloat(b.OpenOrderInitialMargin),
			CrossWalletBalance:     parseFloat(b.CrossWalletBalance),
			CrossUnPnL:             parseFloat(b.CrossUnPnl),
			AvailableBalance:       parseFloat(b.AvailableBalance),
			MaxWithdrawAmount:      parseFloat(b.MaxWithdrawAmount),
		}
	}

	positions := make([]PositionInfo, len(resp.Positions))
	for i, p := range resp.Positions {
		positions[i] = PositionInfo{
			Symbol:                 p.Symbol,
			InitialMargin:          parseFloat(p.InitialMargin),
			MaintMargin:            parseFloat(p.MaintMargin),
			UnrealizedProfit:       parseFloat(p.UnrealizedProfit),
			PositionInitialMargin:  parseFloat(p.PositionInitialMargin),
			OpenOrderInitialMargin: parseFloat(p.OpenOrderInitialMargin),
			Leverage:               parseInt(p.Leverage),
			Isolated:               p.Isolated,
			EntryPrice:             parseFloat(p.EntryPrice),
			MarkPrice:              0, // MarkPrice tidak tersedia di AccountPosition
			PositionSide:           string(p.PositionSide),
			PositionAmt:            parseFloat(p.PositionAmt),
			Notional:               parseFloat(p.Notional),
		}
	}

	return &AccountInfo{
		AvailableBalance:      parseFloat(resp.AvailableBalance),
		TotalBalance:          parseFloat(resp.TotalWalletBalance),
		CrossWalletBalance:    parseFloat(resp.TotalCrossWalletBalance),
		CrossUnPnL:            parseFloat(resp.TotalCrossUnPnl),
		AvailableCrossBalance: parseFloat(resp.AvailableBalance),
		MaxWithdrawAmount:     parseFloat(resp.MaxWithdrawAmount),
		MarginBalance:         parseFloat(resp.TotalMarginBalance),
		WalletBalance:         parseFloat(resp.TotalWalletBalance),
		Balances:              balances,
		Positions:             positions,
	}, nil
}

// GetBalance gets account balance for a specific asset
func (c *Client) GetBalance(asset string) (*BalanceInfo, error) {
	account, err := c.GetAccountInfo()
	if err != nil {
		return nil, err
	}

	for _, b := range account.Balances {
		if b.Asset == asset {
			return &b, nil
		}
	}

	return nil, fmt.Errorf("%w: asset not found: %s", ErrInvalidSymbol, asset)
}

// GetPosition gets position for a symbol
// Cached in Redis for 10 seconds
func (c *Client) GetPosition(symbol string) (*PositionInfo, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("binance:futures:position:%s", symbol)

	// Try cache first
	if c.IsCacheAvailable() {
		var cachedPosition PositionInfo
		err := c.cache.GetJSON(ctx, cacheKey, &cachedPosition)
		if err == nil && cachedPosition.Symbol != "" {
			return &cachedPosition, nil
		}
	}

	// Cache miss - fetch from account info
	account, err := c.GetAccountInfo()
	if err != nil {
		return nil, err
	}

	for _, p := range account.Positions {
		if p.Symbol == symbol {
			// Cache for 10 seconds
			if c.IsCacheAvailable() {
				_ = c.cache.SetJSON(ctx, cacheKey, p, 10*time.Second)
			}
			return &p, nil
		}
	}

	return nil, fmt.Errorf("%w: position not found for symbol: %s", ErrPositionNotFound, symbol)
}

// ============================================================================
// Order Services
// ============================================================================

// PlaceOrder places a new futures order
func (c *Client) PlaceOrder(req *PlaceOrderRequest) (*OrderResponse, error) {
	ctx := context.Background()
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get symbol info for precision adjustment
	symbolInfo, err := c.GetSymbolInfo(req.Symbol)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get symbol info: %v", ErrInvalidSymbol, err)
	}

	// Adjust quantity and price to match exchange precision
	adjustedQuantity := AdjustQuantityPrecision(req.Quantity, symbolInfo.StepSize)
	adjustedPrice := req.Price
	if req.Price > 0 {
		adjustedPrice = AdjustPricePrecision(req.Price, symbolInfo.TickSize)
	}

	// Validate adjusted quantity
	if adjustedQuantity <= 0 {
		return nil, fmt.Errorf("%w: quantity too small after precision adjustment (step size: %f)", ErrInvalidQuantity, symbolInfo.StepSize)
	}

	// Create order service with adjusted values
	orderService := c.apiClient.NewCreateOrderService().
		Symbol(req.Symbol).
		Side(futures.SideType(req.Side)).
		Type(futures.OrderType(req.Type)).
		Quantity(formatQuantity(adjustedQuantity))

	// Add optional parameters
	if adjustedPrice > 0 {
		orderService.Price(formatPrice(adjustedPrice))
	}

	if req.StopPrice > 0 {
		adjustedStopPrice := AdjustPricePrecision(req.StopPrice, symbolInfo.TickSize)
		orderService.StopPrice(formatPrice(adjustedStopPrice))
	}

	if req.TimeInForce != "" {
		orderService.TimeInForce(futures.TimeInForceType(req.TimeInForce))
	}

	if req.PositionSide != "" {
		orderService.PositionSide(futures.PositionSideType(req.PositionSide))
	}

	if req.ReduceOnly {
		orderService.ReduceOnly(req.ReduceOnly)
	}

	if req.WorkingType != "" {
		orderService.WorkingType(futures.WorkingType(req.WorkingType))
	}

	if req.PriceProtect {
		orderService.PriceProtect(req.PriceProtect)
	}

	// Execute order
	resp, err := orderService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to place order: %v", ErrOrderFailed, err)
	}

	return &OrderResponse{
		OrderID:          resp.OrderID,
		Symbol:           resp.Symbol,
		Status:           string(resp.Status),
		ClientOrderID:    resp.ClientOrderID,
		Price:            parseFloat(resp.Price),
		AveragePrice:     parseFloat(resp.AvgPrice),
		OrigQuantity:     parseFloat(resp.OrigQuantity),
		ExecutedQuantity: parseFloat(resp.ExecutedQuantity),
		CumQuantity:      parseFloat(resp.CumQuote),
		CumQuote:         parseFloat(resp.CumQuote),
		TimeInForce:      string(resp.TimeInForce),
		Type:             string(resp.Type),
		Side:             string(resp.Side),
		PositionSide:     string(resp.PositionSide),
		StopPrice:        parseFloat(resp.StopPrice),
		WorkingType:      string(resp.WorkingType),
		OrigType:         string(resp.OrigType),
		UpdateTime:       resp.UpdateTime,
		ReduceOnly:       resp.ReduceOnly,
		ClosePosition:    resp.ClosePosition,
	}, nil
}

// CancelOrder cancels an existing order
func (c *Client) CancelOrder(req *CancelOrderRequest) (*CancelOrderResponse, error) {
	ctx := context.Background()
	cancelService := c.apiClient.NewCancelOrderService().
		Symbol(req.Symbol)

	if req.OrderID > 0 {
		cancelService.OrderID(req.OrderID)
	}

	resp, err := cancelService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to cancel order: %v", ErrOrderFailed, err)
	}

	return &CancelOrderResponse{
		OrderID:          resp.OrderID,
		Symbol:           resp.Symbol,
		Status:           string(resp.Status),
		ClientOrderID:    resp.ClientOrderID,
		Price:            parseFloat(resp.Price),
		AveragePrice:     0, // Tidak tersedia di CancelOrderResponse
		OrigQuantity:     parseFloat(resp.OrigQuantity),
		ExecutedQuantity: parseFloat(resp.ExecutedQuantity),
		CumQuantity:      parseFloat(resp.CumQuantity),
		CumQuote:         parseFloat(resp.CumQuote),
		TimeInForce:      string(resp.TimeInForce),
		Type:             string(resp.Type),
		Side:             string(resp.Side),
		OrigType:         resp.OrigType,
	}, nil
}

// CancelAllOrders cancels all open orders for a symbol
func (c *Client) CancelAllOrders(symbol string) error {
	ctx := context.Background()
	err := c.apiClient.NewCancelAllOpenOrdersService().Symbol(symbol).Do(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to cancel all orders for %s: %v", ErrOrderFailed, symbol, err)
	}
	return nil
}

// GetOrder gets an existing order
func (c *Client) GetOrder(req *GetOrdersRequest) (*OrderResponse, error) {
	ctx := context.Background()
	orderService := c.apiClient.NewGetOrderService().
		Symbol(req.Symbol)

	if req.OrderID > 0 {
		orderService.OrderID(req.OrderID)
	}

	resp, err := orderService.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get order: %v", ErrOrderFailed, err)
	}

	return &OrderResponse{
		OrderID:          resp.OrderID,
		Symbol:           resp.Symbol,
		Status:           string(resp.Status),
		ClientOrderID:    resp.ClientOrderID,
		Price:            parseFloat(resp.Price),
		AveragePrice:     parseFloat(resp.AvgPrice),
		OrigQuantity:     parseFloat(resp.OrigQuantity),
		ExecutedQuantity: parseFloat(resp.ExecutedQuantity),
		CumQuantity:      parseFloat(resp.CumQuantity),
		CumQuote:         parseFloat(resp.CumQuote),
		TimeInForce:      string(resp.TimeInForce),
		Type:             string(resp.Type),
		Side:             string(resp.Side),
		PositionSide:     string(resp.PositionSide),
		StopPrice:        parseFloat(resp.StopPrice),
		WorkingType:      string(resp.WorkingType),
		OrigType:         string(resp.OrigType),
		UpdateTime:       resp.UpdateTime,
		ReduceOnly:       resp.ReduceOnly,
		ClosePosition:    resp.ClosePosition,
	}, nil
}

// GetOpenOrders gets all open orders
func (c *Client) GetOpenOrders(symbol string) ([]OrderResponse, error) {
	ctx := context.Background()
	service := c.apiClient.NewListOpenOrdersService()

	if symbol != "" {
		service.Symbol(symbol)
	}

	resp, err := service.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get open orders: %v", ErrOrderFailed, err)
	}

	orders := make([]OrderResponse, len(resp))
	for i, o := range resp {
		orders[i] = OrderResponse{
			OrderID:          o.OrderID,
			Symbol:           o.Symbol,
			Status:           string(o.Status),
			ClientOrderID:    o.ClientOrderID,
			Price:            parseFloat(o.Price),
			AveragePrice:     parseFloat(o.AvgPrice),
			OrigQuantity:     parseFloat(o.OrigQuantity),
			ExecutedQuantity: parseFloat(o.ExecutedQuantity),
			CumQuantity:      parseFloat(o.CumQuantity),
			CumQuote:         parseFloat(o.CumQuote),
			TimeInForce:      string(o.TimeInForce),
			Type:             string(o.Type),
			Side:             string(o.Side),
			PositionSide:     string(o.PositionSide),
			StopPrice:        parseFloat(o.StopPrice),
			WorkingType:      string(o.WorkingType),
			OrigType:         string(o.OrigType),
			UpdateTime:       o.UpdateTime,
			ReduceOnly:       o.ReduceOnly,
			ClosePosition:    o.ClosePosition,
		}
	}

	return orders, nil
}

// ============================================================================
// Position Services
// ============================================================================

// SetLeverage sets leverage for a symbol
// Leverage is cached in Redis for 3 days
func (c *Client) SetLeverage(req *LeverageRequest) (*LeverageResponse, error) {
	ctx := context.Background()

	// Get symbol info for min/max leverage validation
	symbolInfo, err := c.GetSymbolInfo(req.Symbol)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get symbol info: %v", ErrOrderFailed, err)
	}

	// Validate leverage against symbol limits
	if req.Leverage < symbolInfo.MinLeverage || req.Leverage > symbolInfo.MaxLeverage {
		return nil, fmt.Errorf("%w: leverage must be between %d and %d for %s",
			ErrInvalidLeverage, symbolInfo.MinLeverage, symbolInfo.MaxLeverage, req.Symbol)
	}

	cacheKey := "binance:futures:leverage:" + req.Symbol

	// Check cache first
	if c.IsCacheAvailable() {
		var cachedData struct {
			Leverage    int     `json:"leverage"`
			MaxNotional float64 `json:"max_notional"`
		}

		err := c.cache.GetJSON(ctx, cacheKey, &cachedData)
		if err == nil && cachedData.Leverage == req.Leverage {
			// Same leverage, return from cache
			return &LeverageResponse{
				Symbol:      req.Symbol,
				Leverage:    cachedData.Leverage,
				MaxNotional: cachedData.MaxNotional,
			}, nil
		}
	}

	// Set leverage via API
	resp, err := c.apiClient.NewChangeLeverageService().
		Symbol(req.Symbol).
		Leverage(req.Leverage).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: failed to set leverage: %v", ErrOrderFailed, err)
	}

	// Cache new leverage for 3 days
	if c.IsCacheAvailable() {
		cacheData := struct {
			Leverage    int     `json:"leverage"`
			MaxNotional float64 `json:"max_notional"`
		}{
			Leverage:    req.Leverage,
			MaxNotional: parseFloat(resp.MaxNotionalValue),
		}
		_ = c.cache.SetJSON(ctx, cacheKey, cacheData, 72*time.Hour)
	}

	return &LeverageResponse{
		Symbol:      resp.Symbol,
		Leverage:    resp.Leverage,
		MaxNotional: parseFloat(resp.MaxNotionalValue),
	}, nil
}

// SetMarginMode sets margin mode for a symbol
func (c *Client) SetMarginMode(req *MarginModeRequest) (*MarginModeResponse, error) {
	ctx := context.Background()
	marginMode := futures.MarginTypeIsolated
	if req.MarginMode == 2 {
		marginMode = futures.MarginTypeCrossed
	}

	err := c.apiClient.NewChangeMarginTypeService().
		Symbol(req.Symbol).
		MarginType(marginMode).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: failed to set margin mode: %v", ErrOrderFailed, err)
	}

	return &MarginModeResponse{
		Symbol:     req.Symbol,
		MarginMode: req.MarginMode,
	}, nil
}

// SetPositionSide sets position side mode
func (c *Client) SetPositionSide(req *PositionSideRequest) (*PositionSideResponse, error) {
	ctx := context.Background()
	err := c.apiClient.NewChangePositionModeService().
		DualSide(req.DualSidePosition).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: failed to set position side: %v", ErrOrderFailed, err)
	}

	return &PositionSideResponse{
		DualSidePosition: req.DualSidePosition,
	}, nil
}

// GetPositions gets all positions
func (c *Client) GetPositions() ([]PositionInfo, error) {
	account, err := c.GetAccountInfo()
	if err != nil {
		return nil, err
	}

	// Filter only positions with non-zero amount
	var positions []PositionInfo
	for _, p := range account.Positions {
		if p.PositionAmt != 0 {
			positions = append(positions, p)
		}
	}

	return positions, nil
}

// ClosePosition closes a position by placing an opposite order
func (c *Client) ClosePosition(symbol string, quantity float64, side OrderSide) (*OrderResponse, error) {
	req := &PlaceOrderRequest{
		Symbol:     symbol,
		Side:       side,
		Type:       OrderTypeMarket,
		Quantity:   quantity,
		ReduceOnly: true,
	}

	return c.PlaceOrder(req)
}

// CloseAllPositions closes all open positions
func (c *Client) CloseAllPositions() error {
	positions, err := c.GetPositions()
	if err != nil {
		return err
	}

	for _, p := range positions {
		var side OrderSide
		if p.PositionAmt > 0 {
			side = OrderSideSell
		} else {
			side = OrderSideBuy
		}

		_, err := c.ClosePosition(p.Symbol, absFloat(p.PositionAmt), side)
		if err != nil {
			return fmt.Errorf("failed to close position for %s: %v", p.Symbol, err)
		}
	}

	return nil
}

// absFloat returns absolute value of float64
func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// ============================================================================
// Health Check
// ============================================================================

// Ping tests API connectivity
func (c *Client) Ping() error {
	ctx := context.Background()
	err := c.apiClient.NewPingService().Do(ctx)
	if err != nil {
		return fmt.Errorf("API ping failed: %v", err)
	}
	return nil
}

// GetServerTime gets server time
func (c *Client) GetServerTime() (time.Time, error) {
	ctx := context.Background()
	resp, err := c.apiClient.NewServerTimeService().Do(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get server time: %v", err)
	}
	return time.Unix(0, resp*int64(time.Millisecond)), nil
}

// GetExchangeInfo gets exchange information
// Cached in Redis for 1 week
func (c *Client) GetExchangeInfo() (*futures.ExchangeInfo, error) {
	ctx := context.Background()
	cacheKey := "binance:futures:exchange_info:all"

	// Try cache first
	if c.IsCacheAvailable() {
		var cachedInfo futures.ExchangeInfo
		err := c.cache.GetJSON(ctx, cacheKey, &cachedInfo)
		if err == nil && len(cachedInfo.Symbols) > 0 {
			return &cachedInfo, nil
		}
	}

	// Cache miss - fetch from API
	exchangeInfo, err := c.apiClient.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get exchange info: %v", ErrOrderFailed, err)
	}

	// Cache for 1 week
	if c.IsCacheAvailable() {
		_ = c.cache.SetJSON(ctx, cacheKey, exchangeInfo, 7*24*time.Hour)
	}

	return exchangeInfo, nil
}

// GetSymbolInfo gets trading information for a symbol
func (c *Client) GetSymbolInfo(symbol string) (*SymbolInfo, error) {
	// Get exchange info (handles Redis cache internally)
	exchangeInfo, err := c.GetExchangeInfo()
	if err != nil {
		return nil, err
	}

	// Find symbol info
	var symbolInfo SymbolInfo
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == symbol {
			// Set precision defaults
			symbolInfo.QuantityPrecision = s.QuantityPrecision
			symbolInfo.PricePrecision = s.PricePrecision

			// Parse filters for step size, tick size, and min notional
			for _, filter := range s.Filters {
				switch filter["filterType"] {
				case "LOT_SIZE":
					if stepVal, ok := filter["stepSize"].(string); ok {
						symbolInfo.StepSize, _ = strconv.ParseFloat(stepVal, 64)
					}
				case "PRICE_FILTER":
					if tickVal, ok := filter["tickSize"].(string); ok {
						symbolInfo.TickSize, _ = strconv.ParseFloat(tickVal, 64)
					}
				case "MIN_NOTIONAL":
					if minVal, ok := filter["minNotional"].(string); ok {
						symbolInfo.MinNotional, _ = strconv.ParseFloat(minVal, 64)
					}
				}
			}

			// Set leverage based on symbol type (Binance Futures max leverage by symbol category)
			// Reference: https://www.binance.com/en/futures/leverage-bracket
			symbolInfo.MaxLeverage = getMaxLeverageBySymbol(s.Symbol)
			symbolInfo.MinLeverage = 1
			break
		}
	}

	// Set defaults if not found
	if symbolInfo.MaxLeverage == 0 {
		symbolInfo.MaxLeverage = 125
		symbolInfo.MinLeverage = 1
	}
	if symbolInfo.MinNotional == 0 {
		symbolInfo.MinNotional = 5.0 // Default min notional $5 USDT
	}
	if symbolInfo.QuantityPrecision == 0 {
		symbolInfo.QuantityPrecision = 3
	}
	if symbolInfo.PricePrecision == 0 {
		symbolInfo.PricePrecision = 2
	}
	if symbolInfo.StepSize == 0 {
		symbolInfo.StepSize = 0.001
	}
	if symbolInfo.TickSize == 0 {
		symbolInfo.TickSize = 0.01
	}

	return &symbolInfo, nil
}
