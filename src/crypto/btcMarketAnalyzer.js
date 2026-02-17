import config from "../config.js";
import {
  calculateEMA,
  calculateSMA,
} from "../shared/indicators/movingAverages.js";
import { calculateRSI } from "../shared/indicators/rsi.js";

/**
 * Analyze BTC Market Sentiment (replaces analyzeIHSG for crypto)
 * Uses BTC price action as market-wide sentiment indicator.
 *
 * @param {Array} btcData - OHLCV data for BTCUSDT (daily)
 * @param {string} symbol - The symbol being analyzed (to detect self-reference)
 * @returns {Object} - Market sentiment analysis
 */
export function analyzeBTCMarket(btcData, symbol = "") {
  if (!btcData || btcData.length < 30) {
    return {
      trend: "NEUTRAL",
      strength: 0,
      signal: 0,
      isCrash: false,
      details: ["Insufficient BTC market data"],
    };
  }

  const closes = btcData.map((d) => d.close);
  const curr = closes[closes.length - 1];

  // Calculate indicators
  const ema20 = calculateEMA(closes, 20);
  const ema50 = calculateEMA(closes, 50);
  const rsi = calculateRSI(closes, 14);

  const lastEma20 = ema20[ema20.length - 1];
  const lastEma50 = ema50[ema50.length - 1];
  const lastRsi = rsi[rsi.length - 1] || 50;

  let score = 0;
  const details = [];

  // EMA trend
  if (lastEma20 > lastEma50) {
    score += 30;
    details.push("BTC EMA20 > EMA50");
  } else {
    score -= 30;
    details.push("BTC EMA20 < EMA50");
  }

  // Price vs EMA20
  if (curr > lastEma20) {
    score += 15;
    details.push("BTC above EMA20");
  } else {
    score -= 15;
    details.push("BTC below EMA20");
  }

  // RSI
  if (lastRsi > 50) {
    score += 15;
    details.push(`BTC RSI: ${lastRsi.toFixed(1)}`);
  } else {
    score -= 15;
    details.push(`BTC RSI: ${lastRsi.toFixed(1)}`);
  }

  // 24h change (1-day)
  const prevClose = closes[closes.length - 2];
  const change1d = ((curr - prevClose) / prevClose) * 100;

  // 7-day momentum
  const close7dAgo = closes.length >= 8 ? closes[closes.length - 8] : curr;
  const change7d = ((curr - close7dAgo) / close7dAgo) * 100;

  // Crash detection: BTC drops >5% in 24h
  const crashThreshold = config.CRYPTO?.CRASH_THRESHOLD || -5;
  const isCrash = change1d < crashThreshold;

  if (isCrash) {
    score -= 50;
    details.push(`⚠️ BTC CRASH: ${change1d.toFixed(2)}% (24h)`);
  } else if (change1d > 3) {
    score += 20;
    details.push(`BTC 24h: +${change1d.toFixed(2)}%`);
  } else if (change1d < -2) {
    score -= 20;
    details.push(`BTC 24h: ${change1d.toFixed(2)}%`);
  }

  // 7-day momentum
  if (change7d > 5) {
    score += 15;
    details.push(`BTC 7d: +${change7d.toFixed(2)}%`);
  } else if (change7d < -5) {
    score -= 15;
    details.push(`BTC 7d: ${change7d.toFixed(2)}%`);
  }

  // Self-reference check: if analyzing BTCUSDT, reduce weight since it's redundant
  const isSelf = symbol.toUpperCase().startsWith("BTC");

  return {
    trend: score >= 30 ? "BULLISH" : score <= -30 ? "BEARISH" : "NEUTRAL",
    strength: Math.abs(score),
    signal: Math.min(100, Math.max(-100, score)),
    currentPrice: curr,
    change1d,
    change7d,
    isCrash,
    isSelf,
    details,
  };
}
