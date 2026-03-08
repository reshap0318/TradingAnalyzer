package dtos

// ConfigRequest represents a request to create or update a config
type ConfigRequest struct {
	ConfigKey string `json:"config_key" binding:"required,max=100"`
	Value     string `json:"value" binding:"required"`
	Category  string `json:"category" binding:"required,max=50"`
}

// ConfigData represents config data in responses
type ConfigData struct {
	ID        uint   `json:"id"`
	ConfigKey string `json:"config_key"`
	Value     string `json:"value"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at"`
}
