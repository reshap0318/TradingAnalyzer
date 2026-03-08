package binance

import "errors"

// Common errors for Binance Futures client
var (
	ErrAPIKeyRequired      = errors.New("API key is required")
	ErrSecretKeyRequired   = errors.New("secret key is required")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidSymbol       = errors.New("invalid symbol")
	ErrOrderFailed         = errors.New("order failed")
	ErrAPIRateLimit        = errors.New("API rate limit exceeded")
	ErrInvalidOrderType    = errors.New("invalid order type")
	ErrInvalidPrice        = errors.New("invalid price")
	ErrInvalidQuantity     = errors.New("invalid quantity")
	ErrOrderNotFound       = errors.New("order not found")
	ErrPositionNotFound    = errors.New("position not found")
	ErrInvalidLeverage     = errors.New("invalid leverage")
	ErrInvalidMarginMode   = errors.New("invalid margin mode")
	ErrMarketClosed        = errors.New("market is closed")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrInvalidTimestamp    = errors.New("invalid timestamp")
	ErrSignatureError      = errors.New("signature error")
	ErrInvalidParameter    = errors.New("invalid parameter")
)
