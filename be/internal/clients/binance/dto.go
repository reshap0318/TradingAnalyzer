package binance

import "fmt"

// OrderSide represents order side
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderType represents order type
type OrderType string

const (
	OrderTypeLimit              OrderType = "LIMIT"
	OrderTypeMarket             OrderType = "MARKET"
	OrderTypeStopMarket         OrderType = "STOP_MARKET"
	OrderTypeStopLimit          OrderType = "STOP_LOSS_LIMIT"
	OrderTypeTakeProfit         OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitMarket   OrderType = "TAKE_PROFIT_MARKET"
	OrderTypeTakeProfitLimit    OrderType = "TAKE_PROFIT_LIMIT"
	OrderTypeTrailingStopMarket OrderType = "TRAILING_STOP_MARKET"
)

// PositionSide represents position side
type PositionSide string

const (
	PositionSideBoth  PositionSide = "BOTH"
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

// TimeInForce represents time in force
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
	TimeInForceIOC TimeInForce = "IOC" // Immediate Or Cancel
	TimeInForceFOK TimeInForce = "FOK" // Fill Or Kill
	TimeInForceGTX TimeInForce = "GTX" // Good Till Crossing (post only)
)

// WorkingType represents working type for stop orders
type WorkingType string

const (
	WorkingTypeMarkPrice     WorkingType = "MARK_PRICE"
	WorkingTypeContractPrice WorkingType = "CONTRACT_PRICE"
)

// PlaceOrderRequest represents request to place futures order
type PlaceOrderRequest struct {
	Symbol           string       `json:"symbol" binding:"required"`
	Side             OrderSide    `json:"side" binding:"required,oneof=BUY SELL"`
	Type             OrderType    `json:"type" binding:"required"`
	Quantity         float64      `json:"quantity" binding:"required,gt=0"`
	Price            float64      `json:"price"`
	StopPrice        float64      `json:"stop_price"`
	TimeInForce      TimeInForce  `json:"time_in_force"`
	PositionSide     PositionSide `json:"position_side"`
	ReduceOnly       bool         `json:"reduce_only"`
	NewOrderRespType string       `json:"new_order_resp_type"`
	WorkingType      WorkingType  `json:"working_type"`
	PriceProtect     bool         `json:"price_protect"`
}

// Validate validates the request
func (r *PlaceOrderRequest) Validate() error {
	if r.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if r.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	// Validate price for limit orders
	if r.Type == OrderTypeLimit || r.Type == OrderTypeStopLimit || r.Type == OrderTypeTakeProfitLimit {
		if r.Price <= 0 {
			return fmt.Errorf("price is required for limit orders")
		}
	}

	// Validate time in force for limit orders
	if r.Type == OrderTypeLimit || r.Type == OrderTypeStopLimit || r.Type == OrderTypeTakeProfitLimit {
		if r.TimeInForce == "" {
			return fmt.Errorf("time_in_force is required for limit orders")
		}
	}

	return nil
}

// OrderResponse represents futures order response
type OrderResponse struct {
	OrderID          int64   `json:"order_id"`
	Symbol           string  `json:"symbol"`
	Status           string  `json:"status"`
	ClientOrderID    string  `json:"client_order_id"`
	Price            float64 `json:"price"`
	AveragePrice     float64 `json:"average_price"`
	OrigQuantity     float64 `json:"orig_quantity"`
	ExecutedQuantity float64 `json:"executed_quantity"`
	CumQuantity      float64 `json:"cum_quantity"`
	CumQuote         float64 `json:"cum_quote"`
	TimeInForce      string  `json:"time_in_force"`
	Type             string  `json:"type"`
	Side             string  `json:"side"`
	PositionSide     string  `json:"position_side"`
	StopPrice        float64 `json:"stop_price"`
	WorkingType      string  `json:"working_type"`
	OrigType         string  `json:"orig_type"`
	UpdateTime       int64   `json:"update_time"`
	ActivatePrice    float64 `json:"activate_price"`
	PriceRate        float64 `json:"price_rate"`
	ReduceOnly       bool    `json:"reduce_only"`
	ClosePosition    bool    `json:"close_position"`
}

// AccountInfo represents futures account information
type AccountInfo struct {
	AvailableBalance      float64        `json:"available_balance"`
	TotalBalance          float64        `json:"total_balance"`
	CrossWalletBalance    float64        `json:"cross_wallet_balance"`
	CrossUnPnL            float64        `json:"cross_un_pnl"`
	AvailableCrossBalance float64        `json:"available_cross_balance"`
	MaxWithdrawAmount     float64        `json:"max_withdraw_amount"`
	MarginBalance         float64        `json:"margin_balance"`
	WalletBalance         float64        `json:"wallet_balance"`
	Balances              []BalanceInfo  `json:"balances"`
	Positions             []PositionInfo `json:"positions"`
}

// BalanceInfo represents balance information
type BalanceInfo struct {
	Asset                  string  `json:"asset"`
	WalletBalance          float64 `json:"wallet_balance"`
	UnrealizedProfit       float64 `json:"unrealized_profit"`
	MarginBalance          float64 `json:"margin_balance"`
	MaintMargin            float64 `json:"maint_margin"`
	InitialMargin          float64 `json:"initial_margin"`
	PositionInitialMargin  float64 `json:"position_initial_margin"`
	OpenOrderInitialMargin float64 `json:"open_order_initial_margin"`
	CrossWalletBalance     float64 `json:"cross_wallet_balance"`
	CrossUnPnL             float64 `json:"cross_un_pnl"`
	AvailableBalance       float64 `json:"available_balance"`
	MaxWithdrawAmount      float64 `json:"max_withdraw_amount"`
}

// PositionInfo represents position information
type PositionInfo struct {
	Symbol                 string  `json:"symbol"`
	InitialMargin          float64 `json:"initial_margin"`
	MaintMargin            float64 `json:"maint_margin"`
	UnrealizedProfit       float64 `json:"unrealized_profit"`
	PositionInitialMargin  float64 `json:"position_initial_margin"`
	OpenOrderInitialMargin float64 `json:"open_order_initial_margin"`
	Leverage               int     `json:"leverage"`
	Isolated               bool    `json:"isolated"`
	EntryPrice             float64 `json:"entry_price"`
	MarkPrice              float64 `json:"mark_price"`
	PositionSide           string  `json:"position_side"`
	PositionAmt            float64 `json:"position_amt"`
	Notional               float64 `json:"notional"`
}

// PriceInfo represents price information
type PriceInfo struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Time   int64   `json:"time"`
}

// KlineInfo represents kline/candlestick information
type KlineInfo struct {
	OpenTime                 int64   `json:"open_time"`
	Open                     float64 `json:"open"`
	High                     float64 `json:"high"`
	Low                      float64 `json:"low"`
	Close                    float64 `json:"close"`
	Volume                   float64 `json:"volume"`
	CloseTime                int64   `json:"close_time"`
	QuoteAssetVolume         float64 `json:"quote_asset_volume"`
	TradesCount              int     `json:"trades_count"`
	TakerBuyBaseAssetVolume  float64 `json:"taker_buy_base_asset_volume"`
	TakerBuyQuoteAssetVolume float64 `json:"taker_buy_quote_asset_volume"`
}

// TickerInfo represents 24hr ticker information
type TickerInfo struct {
	Symbol             string  `json:"symbol"`
	PriceChange        float64 `json:"price_change"`
	PriceChangePercent float64 `json:"price_change_percent"`
	WeightedAvgPrice   float64 `json:"weighted_avg_price"`
	LastPrice          float64 `json:"last_price"`
	LastQty            float64 `json:"last_qty"`
	OpenPrice          float64 `json:"open_price"`
	HighPrice          float64 `json:"high_price"`
	LowPrice           float64 `json:"low_price"`
	Volume             float64 `json:"volume"`
	QuoteVolume        float64 `json:"quote_volume"`
	OpenTime           int64   `json:"open_time"`
	CloseTime          int64   `json:"close_time"`
	FirstID            int64   `json:"first_id"`
	LastID             int64   `json:"last_id"`
	Count              int64   `json:"count"`
}

// LeverageRequest represents request to change leverage
type LeverageRequest struct {
	Symbol   string `json:"symbol" binding:"required"`
	Leverage int    `json:"leverage" binding:"required,min=1,max=125"`
}

// LeverageResponse represents leverage response
type LeverageResponse struct {
	Symbol          string  `json:"symbol"`
	Leverage        int     `json:"leverage"`
	MaxNotional     float64 `json:"max_notional"`
	LeverageBracket int     `json:"leverage_bracket"`
}

// CancelOrderRequest represents request to cancel order
type CancelOrderRequest struct {
	Symbol      string `json:"symbol" binding:"required"`
	OrderID     int64  `json:"order_id"`
	OrigClOrdID string `json:"orig_cl_ord_id"`
}

// CancelOrderResponse represents cancel order response
type CancelOrderResponse struct {
	OrderID          int64   `json:"order_id"`
	Symbol           string  `json:"symbol"`
	Status           string  `json:"status"`
	ClientOrderID    string  `json:"client_order_id"`
	Price            float64 `json:"price"`
	AveragePrice     float64 `json:"average_price"`
	OrigQuantity     float64 `json:"orig_quantity"`
	ExecutedQuantity float64 `json:"executed_quantity"`
	CumQuantity      float64 `json:"cum_quantity"`
	CumQuote         float64 `json:"cum_quote"`
	TimeInForce      string  `json:"time_in_force"`
	Type             string  `json:"type"`
	Side             string  `json:"side"`
	OrigType         string  `json:"orig_type"`
}

// GetOrdersRequest represents request to get orders
type GetOrdersRequest struct {
	Symbol    string `json:"symbol" binding:"required"`
	OrderID   int64  `json:"order_id"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Limit     int    `json:"limit"`
}

// MarginModeRequest represents request to change margin mode
type MarginModeRequest struct {
	Symbol     string `json:"symbol" binding:"required"`
	MarginMode int    `json:"margin_mode" binding:"required"` // 1: Isolated, 2: Crossed
}

// MarginModeResponse represents margin mode response
type MarginModeResponse struct {
	Symbol     string `json:"symbol"`
	MarginMode int    `json:"margin_mode"`
}

// PositionSideRequest represents request to change position side
type PositionSideRequest struct {
	DualSidePosition bool `json:"dual_side_position"`
}

// PositionSideResponse represents position side response
type PositionSideResponse struct {
	DualSidePosition bool `json:"dual_side_position"`
}

// PlaceAlgoOrderRequest represents request to place conditional algo order
type PlaceAlgoOrderRequest struct {
	Symbol        string    `json:"symbol" binding:"required"`
	Side          OrderSide `json:"side" binding:"required,oneof=BUY SELL"`
	Type          OrderType `json:"type" binding:"required"`
	Quantity      float64   `json:"quantity"` // Not required if ClosePosition is true
	TriggerPrice  float64   `json:"trigger_price" binding:"required,gt=0"`
	ClosePosition bool      `json:"close_position"`
}

// Validate validates the algo order request
func (r *PlaceAlgoOrderRequest) Validate() error {
	if r.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if r.TriggerPrice <= 0 {
		return fmt.Errorf("trigger_price must be greater than 0")
	}
	if !r.ClosePosition && r.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0 if not closing position")
	}
	return nil
}

// AlgoOrderResponse represents conditional algo order response
type AlgoOrderResponse struct {
	AlgoID       int64   `json:"algoId"`
	ClientAlgoID string  `json:"clientAlgoId"`
	AlgoType     string  `json:"algoType"`
	OrderType    string  `json:"orderType"`
	Symbol       string  `json:"symbol"`
	Side         string  `json:"side"`
	Quantity     string  `json:"quantity"`
	AlgoStatus   string  `json:"algoStatus"`
	TriggerPrice string  `json:"triggerPrice"`
	Code         int64   `json:"code,omitempty"` // Present on errors
	Msg          string  `json:"msg,omitempty"`  // Present on errors
}

type MultiKlineRequest struct {
	Interval string
	Limit    int
}

// SymbolInfo holds trading information for a symbol
type SymbolInfo struct {
	QuantityPrecision int     `json:"quantity_precision"`
	PricePrecision    int     `json:"price_precision"`
	StepSize          float64 `json:"step_size"`
	TickSize          float64 `json:"tick_size"`
	MinLeverage       int     `json:"min_leverage"`
	MaxLeverage       int     `json:"max_leverage"`
	MinNotional       float64 `json:"min_notional"`
}
