package dtos

// StrategyRequest represents a request to create or update a strategy
type StrategyRequest struct {
	StrategyName     string                     `json:"strategy_name" binding:"required,max=100"`
	PrimaryTF        string                     `json:"primary_tf" binding:"required,max=5"`
	IsActive         bool                       `json:"is_active"`
	Timeframes       []StrategyTimeframeRequest `json:"timeframes" binding:"dive"`
	IndicatorWeights []StrategyIndicatorRequest `json:"indicator_weights" binding:"dive"`
	MoneyManagement  []StrategyMoneyMgmtRequest `json:"money_management" binding:"dive"`
}

// StrategyTimeframeRequest represents a timeframe weight in strategy
type StrategyTimeframeRequest struct {
	TF     string  `json:"tf" binding:"required,max=5"`
	Weight float64 `json:"weight" binding:"required,gte=0,lte=1"`
}

// StrategyIndicatorRequest represents an indicator weight in strategy
type StrategyIndicatorRequest struct {
	IndicatorID uint    `json:"indicator_id" binding:"required"`
	Weight      float64 `json:"weight" binding:"required,gte=0,lte=1"`
}

// StrategyMoneyMgmtRequest represents a money management parameter in strategy
type StrategyMoneyMgmtRequest struct {
	Parameter string `json:"parameter" binding:"required,max=50"`
	Value     string `json:"value" binding:"required"`
}

// MMConfigResponse holds money management configuration (same as config.MMConfig)
// This is a duplicate to avoid circular dependency
type MMConfigResponse struct {
	MIN_CONFIDENCE         int8    `json:"min_confidence"`         // minimum confidence level to take a trade
	MAX_DAILY_TRADES       int8    `json:"max_daily_trades"`       // Soft limit trades per hari
	MAX_DAILY_LOSS_PERCENT float32 `json:"max_daily_loss_percent"` // ..% from balance (HARD LIMIT)
	MAX_DAILY_LOSS_COUNT   int8    `json:"max_daily_loss_count"`   //Max consecutive losses (HARD LIMIT)
	RISK_REWARD_RATIO      float32 `json:"risk_reward_ratio"`      // Minimum R:R required
	RISK_REWARD_TARGET     float32 `json:"risk_reward_target"`     // Target R:R untuk excellent setups
	RISK_ENTRY_BUFFER      float32 `json:"risk_entry_buffer"`
	MAX_POSITION_SIZE      float32 `json:"max_position_size"`      // ..% from balance per trade (HARD LIMIT)
	LEVERAGE               int8    `json:"leverage"`               // Leverage to use for futures trading
	IS_AGRESSIVE           bool    `json:"is_agressive"`           // true = hybrid entry, false = conservative
	ORDER_EXPIRATION_HOURS int8    `json:"order_expiration_hours"` // Hours before pending orders expire
}

// StrategyData represents strategy data in responses
type StrategyData struct {
	ID               uint                    `json:"id"`
	StrategyName     string                  `json:"strategy_name"`
	PrimaryTF        string                  `json:"primary_tf"`
	IsActive         bool                    `json:"is_active"`
	CreatedAt        string                  `json:"created_at"`
	UpdatedAt        string                  `json:"updated_at"`
	Timeframes       []StrategyTimeframeData `json:"timeframes"`
	IndicatorWeights []StrategyIndicatorData `json:"indicator_weights"`
	MoneyManagement  *MMConfigResponse       `json:"money_management,omitempty"`
}

// StrategyTimeframeData represents timeframe weight data in responses
type StrategyTimeframeData struct {
	ID              uint           `json:"id"`
	TimeframeName   string         `json:"tf"`
	Weight          float64        `json:"weight"`
	TimeframeDetail *TimeframeData `json:"timeframe_detail,omitempty"`
}

// StrategyIndicatorData represents indicator weight data in responses
type StrategyIndicatorData struct {
	ID              uint           `json:"id"`
	IndicatorID     uint           `json:"indicator_id"`
	Weight          float64        `json:"weight"`
	IndicatorDetail *IndicatorData `json:"indicator_detail,omitempty"`
}
