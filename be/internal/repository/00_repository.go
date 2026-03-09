package repository

import "gorm.io/gorm"

type Repositories struct {
	TxManager           *TransactionManager
	Auth                *AuthRepository
	Threshold           *ThresholdRepository
	Timeframe           *TimeframeRepository
	Watchlist           *WatchlistRepository
	Config              *ConfigRepository
	Indicator           *IndicatorRepository
	Trade               *TradeRepository
	TradeEntry          *TradeEntryRepository
	Backtest            *BacktestRepository
	BacktestTrade       *BacktestTradeRepository
	Strategy            *StrategyRepository
	StrategyTimeframe   *StrategyTimeframeRepository
	StrategyIndicator   *StrategyIndicatorRepository
	StrategyMoneyMgmt   *StrategyMoneyMgmtRepository
}

func NewRepositories(db *gorm.DB) (*Repositories, error) {
	// init repositories that have dependencies first
	tradeEntryRepo := NewTradeEntryRepository(db)
	TxManagerRepo := NewTransactionManager(db)
	ThresholdRepo := NewThresholdRepository(db)
	TimeframeRepo := NewTimeframeRepository(db)
	WatchlistRepo := NewWatchlistRepository(db)
	ConfigRepo := NewConfigRepository(db)
	IndicatorRepo := NewIndicatorRepository(db)
	TradeRepo := NewTradeRepository(db)
	BacktestRepo := NewBacktestRepository(db)
	BacktestTradeRepo := NewBacktestTradeRepository(db)
	
	// Strategy repositories
	StrategyRepo := NewStrategyRepository(db)
	StrategyTimeframeRepo := NewStrategyTimeframeRepository(db)
	StrategyIndicatorRepo := NewStrategyIndicatorRepository(db)
	StrategyMoneyMgmtRepo := NewStrategyMoneyMgmtRepository(db)

	return &Repositories{
		TxManager:           TxManagerRepo,
		Auth:                NewAuthRepository(),
		Threshold:           ThresholdRepo,
		Timeframe:           TimeframeRepo,
		Watchlist:           WatchlistRepo,
		Config:              ConfigRepo,
		Indicator:           IndicatorRepo,
		Trade:               TradeRepo,
		TradeEntry:          tradeEntryRepo,
		Backtest:            BacktestRepo,
		BacktestTrade:       BacktestTradeRepo,
		Strategy:            StrategyRepo,
		StrategyTimeframe:   StrategyTimeframeRepo,
		StrategyIndicator:   StrategyIndicatorRepo,
		StrategyMoneyMgmt:   StrategyMoneyMgmtRepo,
	}, nil
}
