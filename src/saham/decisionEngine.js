import config from "../config.js";
import { analyzeMA } from "../shared/indicators/movingAverages.js";
import { analyzeRSI } from "../shared/indicators/rsi.js";
import { analyzeMACD } from "../shared/indicators/macd.js";
import { analyzeBollingerBands } from "../shared/indicators/bollingerBands.js";
import { analyzeStochastic } from "../shared/indicators/stochastic.js";
import { analyzeVolume } from "../shared/indicators/volume.js";
import { analyzeMultiTimeframe } from "./timeframeManager.js";
import { detectCandlePatterns } from "../shared/indicators/candlePatterns.js";

export function makeDecision(
  ohlcData,
  multiTfData,
  ihsgAnalysis,
  timeframes,
  marketIndexLabel = "IHSG"
) {
  const opens = ohlcData.map((d) => d.open);
  const highs = ohlcData.map((d) => d.high);
  const lows = ohlcData.map((d) => d.low);
  const closes = ohlcData.map((d) => d.close);
  const w = config.SAHAM.WEIGHTS;

  const ind = {
    ma: analyzeMA(closes),
    rsi: analyzeRSI(closes),
    macd: analyzeMACD(closes),
    bb: analyzeBollingerBands(closes),
    stoch: analyzeStochastic(ohlcData),
    volume: analyzeVolume(ohlcData),
  };
  const mtf = analyzeMultiTimeframe(multiTfData, timeframes);

  // Detect Patterns for ALL timeframes
  // Detect Patterns for ALL timeframes (Last 5 candles)
  const patterns = {};
  const HISTORY_COUNT = 5;

  const tfOrder = timeframes;
  tfOrder.forEach((tf) => {
    const data = multiTfData[tf];
    if (!data) return;

    const tfPatterns = [];
    // Iterate from newest (end) to back
    for (let i = 0; i < HISTORY_COUNT; i++) {
      const idx = data.length - 1 - i;
      if (idx < 5) break; // Need at least 5 candles for detection logic

      const slice = data.slice(0, idx + 1);
      const opens = slice.map((d) => d.open);
      const highs = slice.map((d) => d.high);
      const lows = slice.map((d) => d.low);
      const closes = slice.map((d) => d.close);

      const dets = detectCandlePatterns(opens, highs, lows, closes);
      tfPatterns.push({
        date: data[idx].date, // Include date for clarity
        patterns: dets,
      });
    }
    patterns[tf] = tfPatterns;
  });

  // Daily patterns for scoring (Latest candle is at index 0)
  const dailyPatterns = patterns["1D"]?.[0]?.patterns || [];

  let score = 0;

  // Base Indicators
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
      indicator: marketIndexLabel,
      rawSignal: ihsgAnalysis.signal,
      weight: w.IHSG,
      contribution: ihsgAnalysis.signal * w.IHSG,
      trend: ihsgAnalysis.trend,
      details: ihsgAnalysis.details,
    },
  ];

  // Candle Pattern Scoring (Bonus Score)
  let patternScore = 0;
  if (dailyPatterns.length > 0) {
    dailyPatterns.forEach((p) => {
      if (p === "Bullish Engulfing") patternScore += 15;
      if (p === "Morning Star") patternScore += 20;
      if (p === "Hammer") patternScore += 10;
      if (p === "Piercing Line") patternScore += 10;
      if (p === "Bullish Marubozu") patternScore += 15;

      // Bearish Patterns (Negative Score)
      if (p === "Bearish Engulfing") patternScore -= 15;
      if (p === "Evening Star") patternScore -= 20;
      if (p === "Shooting Star") patternScore -= 10;
      if (p === "Dark Cloud Cover") patternScore -= 10;
      if (p === "Bearish Marubozu") patternScore -= 15;

      // Doji is neutral/indecision, no score
    });
    // Add to breakdown logic
    breakdown.push({
      indicator: "Patterns",
      rawSignal: patternScore,
      weight: 1, // Direct score addition
      contribution: patternScore,
      details: dailyPatterns,
    });
    score += patternScore;
  }

  // Trend Regime Bonus (Mimic User Script Logic)
  // Price > EMA200 (+20), Price > EMA50 (+10)
  const lastPrice = closes[closes.length - 1];
  const ema200 = ind.ma.details.find((d) => d.includes("EMA200"))
    ? parseFloat(ind.ma.details.find((d) => d.includes("EMA200")).split(":")[1])
    : null;
  // We don't have exact EMA values in details, relying on MA signal or re-calculating?
  // Actually ind.ma.signal handles it but maybe weight is too low.
  // Let's add explicit Trend Regime check if data available, or boost MA weight.

  // Easier: Boost MA contribution if it's Bullish
  // Current MA implementation checks SMA200, EMA50 etc.

  // Let's reduce IHSG weight impact if NOT crash
  const ihsgWeight = ihsgAnalysis.isCrash ? 0.5 : 0.05; // High impact only if crash
  breakdown.find((b) => b.indicator === marketIndexLabel).weight = ihsgWeight;
  breakdown.find((b) => b.indicator === marketIndexLabel).contribution =
    ihsgAnalysis.signal * ihsgWeight;

  // Add Trend Following Bonus (LuxAlgo Style)
  // If MA and MACD are bullish, we are in a trend.
  // In strong trends, Oscillators (Stoch/RSI) can stay overbought.
  // We should IGNORE Stoch Overbought signal if Trend is Bullish.
  const isUptrend = ind.ma.signal > 0 && ind.macd.signal > 0;

  if (isUptrend) {
    score += 25; // Bonus for alignment (Increased)
    breakdown.push({
      indicator: "Trend Bonus",
      contribution: 25,
      details: ["Strong Uptrend Detected"],
    });

    // Neutralize Stoch Overbought Penalty
    const stochComp = breakdown.find((b) => b.indicator === "Stoch");
    if (stochComp && stochComp.contribution < 0) {
      score -= stochComp.contribution; // Add back the penalty
      stochComp.contribution = 0;
      stochComp.details.push("Overbought Ignored (Trend)");
    }
  }

  // Recalculate Score
  score = 0;
  breakdown.forEach((b) => (score += b.contribution));

  const t = config.SAHAM.THRESHOLDS;
  let signal, strength;

  // Market Crash Override
  if (ihsgAnalysis.isCrash) {
    signal = "WAIT"; // Force Wait/Avoid
    strength = "WEAK";
    score = -50; // Force negative score
    breakdown.push({
      indicator: "CRASH_PROTECTION",
      details: [`${marketIndexLabel} Crash Detected (>1.5% drop)`],
    });
  } else {
    // Standard Decision
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
    ihsg: ihsgAnalysis,
    patterns, // Include patterns in output
  };
}
