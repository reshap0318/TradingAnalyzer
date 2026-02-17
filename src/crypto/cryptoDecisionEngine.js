import config from "../config.js";
import { analyzeMA } from "../shared/indicators/movingAverages.js";
import { analyzeRSI } from "../shared/indicators/rsi.js";
import { analyzeMACD } from "../shared/indicators/macd.js";
import { analyzeBollingerBands } from "../shared/indicators/bollingerBands.js";
import { analyzeStochastic } from "../shared/indicators/stochastic.js";
import { analyzeVolume } from "../shared/indicators/volume.js";
import { detectCandlePatterns } from "../shared/indicators/candlePatterns.js";

/**
 * Analyze a single timeframe for crypto (relaxed data requirement)
 * Falls back gracefully if data is < 200 (MA needs 200 for full analysis)
 */
function analyzeCryptoTimeframe(ohlcData, timeframe) {
  if (!ohlcData || ohlcData.length < 30) {
    return {
      timeframe,
      trend: "NEUTRAL",
      signal: 0,
      details: ["Insufficient data"],
    };
  }

  const closes = ohlcData.map((d) => d.close);

  // MA might return neutral if < 200 data, that's OK
  const ma = analyzeMA(closes);
  const rsi = analyzeRSI(closes);
  const macd = analyzeMACD(closes);

  // Weighted average for timeframe signal
  const avg = ma.signal * 0.35 + rsi.signal * 0.3 + macd.signal * 0.35;

  return {
    timeframe,
    trend: avg >= 25 ? "BULLISH" : avg <= -25 ? "BEARISH" : "NEUTRAL",
    signal: avg,
    indicators: { ma, rsi, macd },
    details: [
      `MA:${ma.trend}`,
      `RSI:${rsi.value?.toFixed(0) || "N/A"}`,
      `MACD:${macd.signal > 0 ? "+" : "-"}`,
    ],
  };
}

/**
 * Multi-timeframe analysis for crypto using crypto-specific weights
 */
function analyzeCryptoMultiTimeframe(multiTfData) {
  const cryptoTfConfig = config.CRYPTO.TIMEFRAMES;
  const results = {};
  let weighted = 0;
  let total = 0;

  for (const [tf, settings] of Object.entries(cryptoTfConfig)) {
    results[tf] = analyzeCryptoTimeframe(multiTfData[tf], tf);
    weighted += results[tf].signal * settings.weight;
    total += settings.weight;
  }

  const trends = Object.values(results).map((r) => r.trend);
  const bull = trends.filter((t) => t === "BULLISH").length;
  const bear = trends.filter((t) => t === "BEARISH").length;

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

/**
 * Crypto Decision Engine — optimized for H1 trading
 *
 * Key differences from stock:
 * - Uses 1H data as primary indicator source
 * - Crypto-specific weights and thresholds
 * - BUY/SELL/WAIT signal output
 * - BTC market sentiment instead of IHSG
 * - No stock-lot rounding
 *
 * @param {Array} hourlyData - 1H OHLCV data (primary)
 * @param {Object} multiTfData - { "15m": [...], "1h": [...], "4h": [...], "1D": [...] }
 * @param {Object} btcMarket - BTC market analysis from btcMarketAnalyzer
 * @returns {Object} - Decision result
 */
export function makeCryptoDecision(hourlyData, multiTfData, btcMarket) {
  if (!hourlyData || hourlyData.length < 30) {
    return {
      signal: "WAIT",
      strength: "NEUTRAL",
      score: 0,
      confidence: 0,
      breakdown: [
        {
          indicator: "DATA",
          contribution: 0,
          details: ["Insufficient 1H data"],
        },
      ],
      indicators: {},
      multiTimeframe: {
        timeframes: {},
        aggregatedSignal: 0,
        alignment: "MIXED",
        details: [],
      },
      btcMarket,
      patterns: {},
    };
  }

  const closes = hourlyData.map((d) => d.close);
  const w = config.CRYPTO.WEIGHTS;

  // --- Primary Indicators (from 1H data) ---
  const ind = {
    ma: analyzeMA(closes),
    rsi: analyzeRSI(closes),
    macd: analyzeMACD(closes),
    bb: analyzeBollingerBands(closes),
    stoch: analyzeStochastic(hourlyData),
    volume: analyzeVolume(hourlyData),
  };

  // --- Multi-Timeframe Analysis ---
  const mtf = analyzeCryptoMultiTimeframe(multiTfData);

  // --- Candle Patterns (all timeframes, last 5 candles) ---
  const patterns = {};
  const HISTORY_COUNT = 5;

  Object.entries(multiTfData).forEach(([tf, data]) => {
    if (!data || data.length < 10) {
      patterns[tf] = [];
      return;
    }
    const tfPatterns = [];
    for (let i = 0; i < HISTORY_COUNT; i++) {
      const idx = data.length - 1 - i;
      if (idx < 5) break;

      const slice = data.slice(0, idx + 1);
      const opens = slice.map((d) => d.open);
      const highs = slice.map((d) => d.high);
      const lows = slice.map((d) => d.low);
      const closes = slice.map((d) => d.close);

      const dets = detectCandlePatterns(opens, highs, lows, closes);
      tfPatterns.push({
        date: data[idx].date,
        patterns: dets,
      });
    }
    patterns[tf] = tfPatterns;
  });

  // 1H patterns for scoring (primary timeframe)
  const hourlyPatterns = patterns["1h"]?.[0]?.patterns || [];

  // --- Score Breakdown ---
  let score = 0;
  const breakdown = [
    {
      indicator: "MA",
      rawSignal: ind.ma.signal,
      weight: w.MA_TREND,
      contribution: ind.ma.signal * w.MA_TREND,
      details: ind.ma.details,
    },
    {
      indicator: "RSI",
      rawSignal: ind.rsi.signal,
      weight: w.RSI,
      contribution: ind.rsi.signal * w.RSI,
      value: ind.rsi.value,
      zone: ind.rsi.zone,
      details: ind.rsi.details,
    },
    {
      indicator: "MACD",
      rawSignal: ind.macd.signal,
      weight: w.MACD,
      contribution: ind.macd.signal * w.MACD,
      details: ind.macd.details,
    },
    {
      indicator: "BB",
      rawSignal: ind.bb.signal,
      weight: w.BOLLINGER,
      contribution: ind.bb.signal * w.BOLLINGER,
      details: ind.bb.details,
    },
    {
      indicator: "Stoch",
      rawSignal: ind.stoch.signal,
      weight: w.STOCHASTIC,
      contribution: ind.stoch.signal * w.STOCHASTIC,
      details: ind.stoch.details,
    },
    {
      indicator: "Vol",
      rawSignal: ind.volume.signal,
      weight: w.VOLUME,
      contribution: ind.volume.signal * w.VOLUME,
      confirmation: ind.volume.confirmation,
      details: ind.volume.details,
    },
    {
      indicator: "MTF",
      rawSignal: mtf.aggregatedSignal,
      weight: w.MULTI_TF,
      contribution: mtf.aggregatedSignal * w.MULTI_TF,
      alignment: mtf.alignment,
      details: mtf.details,
    },
    {
      indicator: "BTC Market",
      rawSignal: btcMarket.signal,
      weight: btcMarket.isSelf ? 0.01 : w.MARKET, // Near-zero weight for BTCUSDT
      contribution: btcMarket.signal * (btcMarket.isSelf ? 0.01 : w.MARKET),
      trend: btcMarket.trend,
      details: btcMarket.details,
    },
  ];

  // --- Candle Pattern Scoring (from 1H timeframe) ---
  let patternScore = 0;
  if (hourlyPatterns.length > 0) {
    hourlyPatterns.forEach((p) => {
      if (p === "Bullish Engulfing") patternScore += 12;
      if (p === "Morning Star") patternScore += 15;
      if (p === "Hammer") patternScore += 8;
      if (p === "Piercing Line") patternScore += 8;
      if (p === "Bullish Marubozu") patternScore += 12;

      if (p === "Bearish Engulfing") patternScore -= 12;
      if (p === "Evening Star") patternScore -= 15;
      if (p === "Shooting Star") patternScore -= 8;
      if (p === "Dark Cloud Cover") patternScore -= 8;
      if (p === "Bearish Marubozu") patternScore -= 12;
    });

    breakdown.push({
      indicator: "Patterns",
      rawSignal: patternScore,
      weight: 1,
      contribution: patternScore,
      details: hourlyPatterns,
    });
  }

  // --- Trend Regime Detection ---
  const isUptrend = ind.ma.signal > 0 && ind.macd.signal > 0;
  const isDowntrend = ind.ma.signal < 0 && ind.macd.signal < 0;

  if (isUptrend) {
    breakdown.push({
      indicator: "Trend Bonus",
      contribution: 20,
      details: ["Strong Uptrend (MA+MACD aligned)"],
    });
    // Neutralize Stoch overbought penalty in uptrend
    const stochComp = breakdown.find((b) => b.indicator === "Stoch");
    if (stochComp && stochComp.contribution < 0) {
      stochComp.contribution = 0;
      stochComp.details.push("Overbought Ignored (Trend)");
    }
  } else if (isDowntrend) {
    breakdown.push({
      indicator: "Trend Bonus",
      contribution: -20,
      details: ["Strong Downtrend (MA+MACD aligned)"],
    });
    // Neutralize Stoch oversold bonus in downtrend
    const stochComp = breakdown.find((b) => b.indicator === "Stoch");
    if (stochComp && stochComp.contribution > 0) {
      stochComp.contribution = 0;
      stochComp.details.push("Oversold Ignored (Trend)");
    }
  }

  // --- Calculate Final Score ---
  score = 0;
  breakdown.forEach((b) => (score += b.contribution));

  // --- Decision ---
  const t = config.CRYPTO.THRESHOLDS;
  let signal, strength;

  // BTC crash override
  if (btcMarket.isCrash && !btcMarket.isSelf) {
    signal = "WAIT";
    strength = "WEAK";
    score = Math.min(score, -30);
    breakdown.push({
      indicator: "CRASH_PROTECTION",
      contribution: 0,
      details: [`BTC Crash Detected (${btcMarket.change1d?.toFixed(2)}%)`],
    });
  } else {
    if (score >= t.STRONG_BUY) {
      signal = "BUY";
      strength = "STRONG";
    } else if (score >= t.BUY) {
      signal = "BUY";
      strength = "MODERATE";
    } else if (score <= t.STRONG_SELL) {
      signal = "SELL";
      strength = "STRONG";
    } else if (score <= t.SELL) {
      signal = "SELL";
      strength = "MODERATE";
    } else {
      signal = "WAIT";
      strength = "NEUTRAL";
    }
  }

  return {
    signal,
    strength,
    score,
    confidence: Math.min(100, Math.abs(score)),
    breakdown,
    indicators: ind,
    multiTimeframe: mtf,
    btcMarket,
    patterns,
  };
}
