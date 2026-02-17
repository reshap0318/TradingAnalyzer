/**
 * Candle Pattern Detection Module
 * Detects common reversal patterns: Bullish Engulfing, Morning Star, Hammer, Piercing Line, Doji
 */

const isGreen = (o, c) => c > o;
const isRed = (o, c) => c < o;
const bodySize = (o, c) => Math.abs(c - o);
const upperWick = (h, o, c) => h - Math.max(o, c);
const lowerWick = (l, o, c) => Math.min(o, c) - l;
const totalRange = (h, l) => h - l;

// 1. Bullish Engulfing
// Prev: Red, Curr: Green. Curr Body engulfs Prev Body.
const isBullishEngulfing = (prev, curr) => {
  return (
    isRed(prev.open, prev.close) &&
    isGreen(curr.open, curr.close) &&
    curr.open <= prev.close && // Open lower/equal to prev close
    curr.close >= prev.open // Close higher/equal to prev open
  );
};

// 2. Hammer
// Small body at top, Long lower wick (> 2x body), Small/No upper wick
const isHammer = (curr) => {
  const body = bodySize(curr.open, curr.close);
  const uWick = upperWick(curr.high, curr.open, curr.close);
  const lWick = lowerWick(curr.low, curr.open, curr.close);

  return (
    lWick >= 2 * body && uWick <= body * 1.0 // Relaxed from 0.5 to 1.0 to capture "imperfect" hammers
  );
};

// 3. Morning Star (3 candles)
// 1: Long Red. 2: Small body (gap down preferred). 3: Strong Green (gap up preferred, closes > midpoint of 1)
const isMorningStar = (c1, c2, c3) => {
  const midpoint = (c1.open + c1.close) / 2;
  return (
    isRed(c1.open, c1.close) &&
    Math.abs(c1.open - c1.close) > totalRange(c1.high, c1.low) * 0.5 && // Long red
    bodySize(c2.open, c2.close) < bodySize(c1.open, c1.close) * 0.3 && // Small c2
    isGreen(c3.open, c3.close) &&
    c3.close > midpoint // Closes above midpoint of c1
  );
};

// 4. Piercing Line
// Prev: Long Red. Curr: Open below prev low, Close > 50% of prev body.
const isPiercingLine = (prev, curr) => {
  const midpoint = (prev.open + prev.close) / 2;
  return (
    isRed(prev.open, prev.close) &&
    curr.open < prev.low && // Gap down open
    isGreen(curr.open, curr.close) &&
    curr.close > midpoint && // Close above midpoint
    curr.close < prev.open // But below open
  );
};

// 5. Doji
// Body very small relative to range
const isDoji = (curr) => {
  return (
    bodySize(curr.open, curr.close) <= totalRange(curr.high, curr.low) * 0.1
  );
};

// 6. Bearish Engulfing
const isBearishEngulfing = (prev, curr) => {
  return (
    isGreen(prev.open, prev.close) &&
    isRed(curr.open, curr.close) &&
    curr.open >= prev.close &&
    curr.close <= prev.open
  );
};

// 7. Shooting Star (Bearish Hammer)
const isShootingStar = (curr) => {
  const body = bodySize(curr.open, curr.close);
  const uWick = upperWick(curr.high, curr.open, curr.close);
  const lWick = lowerWick(curr.low, curr.open, curr.close);
  return uWick >= 2 * body && lWick <= body * 0.5;
};

// 8. Marubozu (Strong Trend)
// Big body, very small wicks
const isMarubozu = (curr) => {
  const body = bodySize(curr.open, curr.close);
  const range = totalRange(curr.high, curr.low);
  return range > 0 && body / range > 0.85; // Body is > 85% of total range
};

// 9. Evening Star (Bearish Morning Star)
const isEveningStar = (c1, c2, c3) => {
  const midpoint = (c1.open + c1.close) / 2;
  return (
    isGreen(c1.open, c1.close) &&
    Math.abs(c1.open - c1.close) > totalRange(c1.high, c1.low) * 0.5 &&
    bodySize(c2.open, c2.close) < bodySize(c1.open, c1.close) * 0.3 &&
    isRed(c3.open, c3.close) &&
    c3.close < midpoint
  );
};

// 10. Dark Cloud Cover (Bearish Piercing)
const isDarkCloudCover = (prev, curr) => {
  const midpoint = (prev.open + prev.close) / 2;
  return (
    isGreen(prev.open, prev.close) &&
    curr.open > prev.high && // Gap up open
    isRed(curr.open, curr.close) &&
    curr.close < midpoint && // Close below midpoint
    curr.close > prev.open
  );
};

/**
 * Detect patterns in the recent data
 * @param {Array} opens
 * @param {Array} highs
 * @param {Array} lows
 * @param {Array} closes
 * @returns {Array.string} Detected patterns
 */
export function detectCandlePatterns(opens, highs, lows, closes) {
  if (closes.length < 5) return [];

  const len = closes.length;
  const getCandle = (i) => ({
    open: opens[i],
    high: highs[i],
    low: lows[i],
    close: closes[i],
  });

  const curr = getCandle(len - 1);
  const prev1 = getCandle(len - 2);
  const prev2 = getCandle(len - 3);

  // Ignore flat candles (no trading activity or error)
  if (totalRange(curr.high, curr.low) === 0) return [];

  const patterns = [];

  // Bullish
  if (isBullishEngulfing(prev1, curr)) patterns.push("Bullish Engulfing");
  if (isHammer(curr)) patterns.push("Hammer");
  if (isMorningStar(prev2, prev1, curr)) patterns.push("Morning Star");
  if (isPiercingLine(prev1, curr)) patterns.push("Piercing Line");

  // Bearish
  if (isBearishEngulfing(prev1, curr)) patterns.push("Bearish Engulfing");
  if (isShootingStar(curr)) patterns.push("Shooting Star");
  if (isEveningStar(prev2, prev1, curr)) patterns.push("Evening Star");
  if (isDarkCloudCover(prev1, curr)) patterns.push("Dark Cloud Cover");

  // Neutral/Trend
  if (isDoji(curr)) patterns.push("Doji");
  if (isMarubozu(curr)) {
    if (isGreen(curr.open, curr.close)) patterns.push("Bullish Marubozu");
    else patterns.push("Bearish Marubozu");
  }

  return patterns;
}
