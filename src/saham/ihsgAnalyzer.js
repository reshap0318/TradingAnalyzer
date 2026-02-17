import config from "../config.js";
import {
  calculateEMA,
  calculateSMA,
} from "../shared/indicators/movingAverages.js";
import { calculateRSI } from "../shared/indicators/rsi.js";

export function analyzeIHSG(ihsgData) {
  if (!ihsgData || ihsgData.length < 50)
    return {
      trend: "NEUTRAL",
      strength: 0,
      signal: 0,
      details: ["Insufficient Market Data"],
    };

  const closes = ihsgData.map((d) => d.close);
  const curr = closes[closes.length - 1];
  const ema20 = calculateEMA(closes, 20),
    ema50 = calculateEMA(closes, 50),
    sma200 = calculateSMA(closes, 200);
  const rsi = calculateRSI(closes, 14);
  const lastEma20 = ema20[ema20.length - 1],
    lastEma50 = ema50[ema50.length - 1];
  const lastSma200 = sma200.length > 0 ? sma200[sma200.length - 1] : curr;
  const lastRsi = rsi[rsi.length - 1] || 50;

  let score = 0;
  const details = [];
  if (lastEma20 > lastEma50) {
    score += 30;
    details.push("EMA20 > EMA50");
  } else {
    score -= 30;
    details.push("EMA20 < EMA50");
  }
  if (curr > lastSma200) {
    score += 25;
    details.push("Above SMA200");
  } else {
    score -= 25;
    details.push("Below SMA200");
  }
  const change5d =
    ((curr - closes[closes.length - 6]) / closes[closes.length - 6]) * 100;

  const prevClose = closes[closes.length - 2];
  const change1d = ((curr - prevClose) / prevClose) * 100;
  const isCrash = change1d < -1.5;

  if (isCrash) {
    score -= 50; // Heavy penalty for crash
    details.push(`⚠️ MARKET CRASH: ${change1d.toFixed(2)}%`);
  }

  if (change5d > 1) {
    score += 20;
    details.push(`5-day: +${change5d.toFixed(2)}%`);
  } else if (change5d < -1) {
    score -= 20;
    details.push(`5-day: ${change5d.toFixed(2)}%`);
  }
  if (lastRsi > 50) {
    score += 15;
    details.push(`RSI: ${lastRsi.toFixed(1)}`);
  } else {
    score -= 15;
    details.push(`RSI: ${lastRsi.toFixed(1)}`);
  }

  return {
    trend: score >= 40 ? "BULLISH" : score <= -40 ? "BEARISH" : "NEUTRAL",
    strength: Math.abs(score),
    signal: Math.min(100, Math.max(-100, score)),
    currentPrice: curr,
    change1d,
    change5d,
    isCrash,
    details,
  };
}

export function calculateCorrelation(stockData, ihsgData, period = 30) {
  if (stockData.length < period || ihsgData.length < period) return 0;
  const sr = [],
    ir = [];
  for (let i = 1; i < period; i++) {
    sr.push(
      (stockData[stockData.length - period + i].close -
        stockData[stockData.length - period + i - 1].close) /
        stockData[stockData.length - period + i - 1].close
    );
    ir.push(
      (ihsgData[ihsgData.length - period + i].close -
        ihsgData[ihsgData.length - period + i - 1].close) /
        ihsgData[ihsgData.length - period + i - 1].close
    );
  }
  const n = sr.length,
    sumX = sr.reduce((a, b) => a + b, 0),
    sumY = ir.reduce((a, b) => a + b, 0);
  const sumXY = sr.reduce((acc, x, i) => acc + x * ir[i], 0),
    sumX2 = sr.reduce((acc, x) => acc + x * x, 0),
    sumY2 = ir.reduce((acc, y) => acc + y * y, 0);
  const denom = Math.sqrt(
    (n * sumX2 - sumX * sumX) * (n * sumY2 - sumY * sumY)
  );
  return denom === 0 ? 0 : (n * sumXY - sumX * sumY) / denom;
}
