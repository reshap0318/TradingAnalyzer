package dtos

// ThresholdRequest represents a request to create or update a threshold
type ThresholdRequest struct {
	Category     string `json:"category" binding:"required,max=20"`
	MinValue     int    `json:"min_value" binding:"required"`
	MaxValue     int    `json:"max_value" binding:"required"`
	Action       string `json:"action" binding:"required,max=10"`
	Color        string `json:"color" binding:"required,max=20"`
	OrderDisplay int    `json:"order_display" binding:"required"`
}

// ThresholdData represents threshold data in responses
type ThresholdData struct {
	ID           uint   `json:"id"`
	Category     string `json:"category"`
	MinValue     int    `json:"min_value"`
	MaxValue     int    `json:"max_value"`
	Action       string `json:"action"`
	Color        string `json:"color"`
	OrderDisplay int    `json:"order_display"`
	CreatedAt    string `json:"created_at"`
}
