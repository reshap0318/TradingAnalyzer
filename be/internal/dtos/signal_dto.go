package dtos

import "time"

// ============== REQUEST DTOs ==============

// SignalAnalyzeRequest represents the request for signal analysis
type SignalAnalyzeRequest struct {
	Symbol     string  `json:"symbol" binding:"required"`
	StrategyID uint    `json:"strategy_id"` // Optional: if not provided, uses active strategy
	Capital    float64 `json:"capital"`     // Optional: capital for analysis (default: $50)
}

// SignalAnalyzeResponse represents the response from signal analysis
type SignalAnalyzeResponse struct {
	Symbol           string           `json:"symbol"`
	PrimaryTimeframe string           `json:"primary_timeframe"`
	Timestamp        time.Time        `json:"timestamp"`
	Signal           SignalInfo       `json:"signal,omitempty"`
	Scoring          ScoringBreakdown `json:"scoring"`
}

type SignalInfo struct {
	Valid       bool         `json:"valid"`
	Signal      string       `json:"signal"`
	CurentPrice float64      `json:"current_price,omitempty"`
	TradingPlan *TradingPlan `json:"trading_plan,omitempty"`
}

// TradingPlan represents TP/SL and risk management (multi-entry support)
type TradingPlan struct {
	Mode    string             `json:"mode"`    // "CONSERVATIVE" or "AGGRESSIVE"
	Entries []TradingPlanEntry `json:"entries"` // Array of entries

	// TP/SL (single for all entries)
	TakeProfit      float64 `json:"take_profit"`
	StopLoss        float64 `json:"stop_loss"`
	Resistance      float64 `json:"resistance"`
	Support         float64 `json:"support"`
	RiskRewardRatio float64 `json:"risk_reward_ratio"`
	BufferPercent   float64 `json:"buffer_percent"`

	// Summary (pre-calculated data for easy access)
	Summary *TradingPlanSummary `json:"summary,omitempty"`
}

// TradingPlanSummary contains pre-calculated summary data for the trading plan
type TradingPlanSummary struct {
	// Position Info
	TotalEntries       int     `json:"total_entries"`        // Number of entries
	TotalPositionValue float64 `json:"total_position_value"` // Total USDT to be used (from capital)
	TotalPositionQty   float64 `json:"total_position_qty"`   // Total coin quantity
	AvgEntryPrice      float64 `json:"avg_entry_price"`      // Weighted average entry price

	// Risk Info (calculated from position value, not capital)
	MaxRiskUSDT     float64 `json:"max_risk_usdt"`     // Max loss in USDT if SL hit
	MaxRiskPercent  float64 `json:"max_risk_percent"`  // Max risk as % of position value
	RiskFromCapital float64 `json:"risk_from_capital"` // Max risk as % of trading capital

	// Profit Info (calculated from position value, not capital)
	TargetProfitUSDT    float64 `json:"target_profit_usdt"`    // Target profit in USDT if TP hit
	TargetProfitPercent float64 `json:"target_profit_percent"` // Target profit as % of position value
	ProfitFromCapital   float64 `json:"profit_from_capital"`   // Target profit as % of trading capital
	EffectiveLeverage   float64 `json:"effective_leverage"`    // Actual leverage used (position_value / capital)
}

type TradingPlanEntry struct {
	EntryNumber   int     `json:"entry_number"`   // 1, 2, 3, ...
	EntryPrice    float64 `json:"entry_price"`    // Harga entry
	PositionSize  string  `json:"position_size"`  // "50%", "30%", "100%"
	PositionValue float64 `json:"position_value"` // USDT amount
	PositionQty   float64 `json:"position_qty"`   // Coin amount
}

// ScoringBreakdown represents scoring breakdown
type ScoringBreakdown struct {
	TotalScore float64               `json:"totalScore"`
	Confidence float64               `json:"confidence"`
	Breakdown  []TimeframeSignalData `json:"breakdown"`
}

// TimeframeSignalData represents data for a single timeframe
type TimeframeSignalData struct {
	Timeframe    string               `json:"timeframe"`
	Trend        string               `json:"trend"`
	RawSignal    float64              `json:"rawSignal"`
	Weight       float64              `json:"weight"`
	Contribution float64              `json:"contribution"`
	Indicator    []IndicatorBreakdown `json:"indicator"`
}

// IndicatorBreakdown represents breakdown for a single indicator
type IndicatorBreakdown struct {
	Name         string      `json:"name"`
	RawSignal    int         `json:"rawSignal"`
	Weight       float64     `json:"weight"`
	Contribution float64     `json:"contribution"`
	Details      []string    `json:"details"`
	Value        interface{} `json:"value,omitempty"`
	Zone         string      `json:"zone,omitempty"`
}

// ============== Signal Raw Data DTOs ==============
// SignalRawRequest represents the request for raw signal data
// Request: symbol + array of timeframes with name and limit
type SignalRawRequest struct {
	Symbol     string                   `json:"symbol" binding:"required"`
	Timeframes []SignalTimeframeRequest `json:"timeframes" binding:"required,dive"`
}

// SignalTimeframeRequest represents a single timeframe request
type SignalTimeframeRequest struct {
	Name  string `json:"name" binding:"required"`  // e.g., "15m", "1h", "4h"
	Limit int    `json:"limit" binding:"required"` // e.g., 100
}

// SignalRawResponse represents the response for raw signal data
// Response: array of timeframe data with raws
type SignalRawResponse struct {
	Symbol     string             `json:"symbol"`
	Timeframes []TimeframeRawData `json:"timeframes"`
}

// TimeframeRawData represents raw data for a single timeframe
type TimeframeRawData struct {
	Timeframe string    `json:"timeframe"` // e.g., "15m", "1h"
	Raws      []RawData `json:"raws"`      // Array of raw OHLCV data
}

// RawData represents a single raw OHLCV data point
type RawData struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}
