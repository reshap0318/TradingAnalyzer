/**
 * Configuration for Trading Analyzer
 * Separated into SAHAM (stock) and CRYPTO sections
 * Shared settings (INDICATORS, SR_SETTINGS, SERVER) at top level
 */
export const TIMEFRAME_MAP = {
  // Scalping (sangat cepat, 1-15 menit hold)
  "1m": ["1m", "3m", "5m", "15m"],
  "3m": ["1m", "3m", "5m", "15m"],
  "5m": ["1m", "5m", "15m", "1h"],

  // Short-term scalping (15-30 menit hold)
  "15m": ["5m", "15m", "1h", "4h"],
  "30m": ["15m", "30m", "1h", "4h"],

  // Day trading (beberapa jam hold)
  "1h": ["15m", "1h", "4h", "1d"],
  "2h": ["30m", "1h", "2h", "4h"],
  "4h": ["1h", "4h", "1d", "1w"],
  "6h": ["1h", "4h", "6h", "1d"],
  "8h": ["4h", "8h", "1d", "1w"],
  "12h": ["4h", "12h", "1d", "1w"],

  // Swing trading (beberapa hari hold)
  "1d": ["4h", "1d", "3d", "1w"],
  "3d": ["1d", "3d", "1w", "1M"],

  // Position trading (mingguan/bulanan)
  "1w": ["1d", "1w", "1M"],
  "1M": ["1w", "1M"],
};

export default {
  DEFAULT_INTERVAL: "1d",
  // ─── Shared Settings (used by both saham & crypto) ───
  INDICATORS: {
    MA: { SMA_PERIODS: [20, 50, 200], EMA_PERIODS: [12, 26] },
    RSI: { PERIOD: 14, OVERBOUGHT: 70, OVERSOLD: 30 },
    MACD: { FAST: 12, SLOW: 26, SIGNAL: 9 },
    BOLLINGER: { PERIOD: 20, STD_DEV: 2 },
    STOCHASTIC: {
      K_PERIOD: 14,
      D_PERIOD: 3,
      SMOOTH: 3,
      OVERBOUGHT: 80,
      OVERSOLD: 20,
    },
    ATR: { PERIOD: 14 },
    VOLUME: { MA_PERIOD: 20 },
  },

  SR_SETTINGS: { LOOKBACK_PERIODS: 100, TOLERANCE: 0.005, MIN_TOUCHES: 2 },

  // ─── Auto Trading System ───
  AUTO_TRADING: {
    ENABLED: false, // Master switch
    MODE: "paper", // "paper" | "live"
    USE_TESTNET:
      process.env.USE_TESTNET === "true" || process.env.USE_TESTNET === "1", // Driven by .env now
    SAFETY: {
      MIN_CONFIDENCE: 70, // Minimum confidence score untuk eksekusi trade
      MAX_DAILY_TRADES: 5, // Menghentikan bot setelah eksekusi trade ke-5
      MAX_DRAWDOWN_PERCENT: 10, // Stop sistem jika rugi di atas 10%
      MIN_RISK_REWARD: 1.5, // Rasio RisktoReward minimum
      BLOCK_ON_CRASH: true,
    },
    ORDER_TYPE: "MARKET", // "MARKET" | "LIMIT"
    DEFAULT_LEVERAGE: 5,
    AUTO_SET_TP_SL: true, // Auto place TP/SL orders
  },

  // ─── Saham (Stock) Specific ───
  SAHAM: {
    DEFAULT_CAPITAL: 10_000_000, // 10 juta IDR
    IHSG_SYMBOL: "^JKSE",

    TIMEFRAMES: {
      "15m": { interval: "15m", period: "5d", weight: 0.15 },
      "1h": { interval: "60m", period: "1mo", weight: 0.25 },
      "4h": { interval: "60m", period: "3mo", weight: 0.3 },
      "1D": { interval: "1d", period: "6mo", weight: 0.3 },
    },

    WEIGHTS: {
      MA_TREND: 0.12,
      RSI: 0.12,
      MACD: 0.18,
      BOLLINGER: 0.12,
      STOCHASTIC: 0.08,
      VOLUME: 0.08,
      MULTI_TF: 0.15,
      IHSG: 0.15,
    },

    THRESHOLDS: {
      STRONG_BUY: 70,
      BUY: 50,
      WAIT_UPPER: 50,
      WAIT_LOWER: -50,
      SELL: -50,
      STRONG_SELL: -70,
    },

    MONEY_MANAGEMENT: {
      MAX_RISK_PER_TRADE: 0.02,
      MAX_POSITION_SIZE: 0.1,
      RISK_REWARD_MIN: 1.5,
      TRAILING_STOP_ACTIVATION: 1.5,
      TRAILING_STOP_DISTANCE: 0.5,
    },

    // IHSG crash = >1.5% drop
    CRASH_THRESHOLD: -1.5,
    // Max SL for stocks = 6%
    MAX_SL_PERCENT: 6,
  },

  // ─── Crypto Specific ───
  CRYPTO: {
    DEFAULT_CAPITAL: 50, // 50 USD
    TIMEFRAMES: {
      "15m": { weight: 0.1 },
      "1h": { weight: 0.35 }, // Primary entry timeframe
      "4h": { weight: 0.3 },
      "1D": { weight: 0.25 },
    },

    WEIGHTS: {
      MA_TREND: 0.12,
      RSI: 0.15, // Higher for crypto momentum
      MACD: 0.2, // Higher for crypto trends
      BOLLINGER: 0.12,
      STOCHASTIC: 0.08,
      VOLUME: 0.1, // Slightly higher (volume spikes matter in crypto)
      MULTI_TF: 0.18, // Multi-TF alignment is key for H1 trading
      MARKET: 0.05, // BTC sentiment (low weight — not IHSG)
    },

    THRESHOLDS: {
      STRONG_BUY: 70,
      BUY: 45,
      WAIT_UPPER: 45,
      WAIT_LOWER: -45,
      SELL: -45,
      STRONG_SELL: -70,
    },

    MONEY_MANAGEMENT: {
      MAX_RISK_PER_TRADE: 0.02,
      MAX_POSITION_SIZE: 0.15,
      RISK_REWARD_MIN: 1.5,
      TRAILING_STOP_ACTIVATION: 2.0,
      TRAILING_STOP_DISTANCE: 0.8,
    },

    // BTC crash = >5% drop in 24h
    CRASH_THRESHOLD: -5,
    // Max SL for crypto = 8%
    MAX_SL_PERCENT: 8,

    // Futures-specific settings
    FUTURES: {
      DEFAULT_LEVERAGE: 5,
      MAX_LEVERAGE: 20,
      MAINTENANCE_MARGIN_RATE: 0.005, // 0.5% (Binance tier 1)
      TAKER_FEE: 0.0004, // 0.04%
      MAKER_FEE: 0.0002, // 0.02%
      FUNDING_INTERVAL_HOURS: 8,
      // Max position as % of capital (adjusted for leverage risk)
      MAX_POSITION_PERCENT: 0.1, // 10% of capital per trade
    },
  },
};
