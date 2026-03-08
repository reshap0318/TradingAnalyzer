package dtos

// TimeframeRequest represents a timeframe in create/update interval request
type TimeframeRequest struct {
	Name      string `json:"name" binding:"required,max=5"`
	InMinutes int    `json:"in_minutes" binding:"required,gt=0"`
}

// TimeframeData represents timeframe data in responses
type TimeframeData struct {
	Name      string `json:"name"`
	InMinutes int    `json:"in_minutes"`
	CreatedAt string `json:"created_at"`
}
