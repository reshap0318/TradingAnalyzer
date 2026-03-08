package dtos

// IndicatorRequest represents a request to create or update a Indicator
type IndicatorRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Description string  `json:"description"`
	Indicator   string  `json:"indicator" binding:"required,max=50"`
	Params      *string `json:"params"` // JSON string
	IsActive    bool    `json:"is_active"`
	Weight      float64 `json:"weight" binding:"required"`
	OrderView   int     `json:"order_view"`
}

// IndicatorInfo represents information about an indicator
type IndicatorInfo struct {
	Key           string      `json:"key"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	DefaultParams interface{} `json:"default_params"`
}

// IndicatorData represents Indicator data in responses
type IndicatorData struct {
	ID            uint           `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Indicator     string         `json:"indicator"`
	Params        interface{}    `json:"params"`
	IsActive      bool           `json:"is_active"`
	Weight        float64        `json:"weight"`
	OrderView     int            `json:"order_view"`
	CreatedAt     string         `json:"created_at"`
	IndicatorInfo *IndicatorInfo `json:"indicator_info,omitempty"`
}
