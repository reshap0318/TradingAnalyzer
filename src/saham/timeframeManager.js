import config from "../config.js";
import { analyzeMA } from "../shared/indicators/movingAverages.js";
import { analyzeRSI } from "../shared/indicators/rsi.js";
import { analyzeMACD } from "../shared/indicators/macd.js";

export function analyzeTimeframe(ohlcData, timeframe) {
  if (!ohlcData || ohlcData.length < 50)
    return {
      timeframe,
      trend: "NEUTRAL",
      signal: 0,
      details: ["Insufficient data"],
    };
  const closes = ohlcData.map((d) => d.close);
  const ma = analyzeMA(closes),
    rsi = analyzeRSI(closes),
    macd = analyzeMACD(closes);
  const avg = ma.signal * 0.4 + rsi.signal * 0.3 + macd.signal * 0.3;
  return {
    timeframe,
    trend: avg >= 30 ? "BULLISH" : avg <= -30 ? "BEARISH" : "NEUTRAL",
    signal: avg,
    details: [
      `MA:${ma.trend}`,
      `RSI:${rsi.value?.toFixed(0) || "N/A"}`,
      `MACD:${macd.signal > 0 ? "+" : "-"}`,
    ],
  };
}

export function analyzeMultiTimeframe(multiTfData, timeframes) {
  const genericWeights = [0.15, 0.25, 0.3, 0.3];
  const results = {};
  let weighted = 0,
    total = 0;

  timeframes.forEach((tf, index) => {
    const data = multiTfData[tf];
    if (data) {
      results[tf] = analyzeTimeframe(data, tf);
      const weight = genericWeights[index] || 0.15;
      weighted += results[tf].signal * weight;
      total += weight;
    }
  });
  const trends = Object.values(results).map((r) => r.trend);
  const bull = trends.filter((t) => t === "BULLISH").length,
    bear = trends.filter((t) => t === "BEARISH").length;
  const alignment =
    bull >= 3
      ? "BULLISH_ALIGNED"
      : bear >= 3
      ? "BEARISH_ALIGNED"
      : bull >= 2 && !bear
      ? "MOSTLY_BULLISH"
      : bear >= 2 && !bull
      ? "MOSTLY_BEARISH"
      : "MIXED";
  return {
    timeframes: results,
    aggregatedSignal: total > 0 ? weighted / total : 0,
    alignment,
    details: Object.entries(results).map(([tf, r]) => `${tf}:${r.trend}`),
  };
}
