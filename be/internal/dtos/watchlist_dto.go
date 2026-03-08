package dtos

// WatchlistRequest represents a request to create or update a watchlist
type WatchlistRequest struct {
	Symbol   string `json:"symbol" binding:"required,max=20"`
	IsActive bool   `json:"is_active"`
}

// WatchlistData represents watchlist data in responses
type WatchlistData struct {
	ID        uint   `json:"id"`
	Symbol    string `json:"symbol"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}
