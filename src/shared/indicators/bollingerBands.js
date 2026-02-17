import config from "../../config.js";
import { calculateSMA } from "./movingAverages.js";

export function calculateBollingerBands(closes) {
  const { PERIOD, STD_DEV } = config.INDICATORS.BOLLINGER;
  if (closes.length < PERIOD)
    return { upper: [], middle: [], lower: [], bandwidth: [] };

  const middle = calculateSMA(closes, PERIOD);
  const upper = [],
    lower = [],
    bandwidth = [];
  for (let i = PERIOD - 1; i < closes.length; i++) {
    const slice = closes.slice(i - PERIOD + 1, i + 1);
    const avg = middle[i - PERIOD + 1];
    const stdDev = Math.sqrt(
      slice.map((v) => Math.pow(v - avg, 2)).reduce((a, b) => a + b, 0) / PERIOD
    );
    upper.push(avg + stdDev * STD_DEV);
    lower.push(avg - stdDev * STD_DEV);
    bandwidth.push(
      ((upper[upper.length - 1] - lower[lower.length - 1]) / avg) * 100
    );
  }
  return { upper, middle, lower, bandwidth };
}

export function analyzeBollingerBands(closes) {
  const bb = calculateBollingerBands(closes);
  if (bb.upper.length < 5)
    return { signal: 0, position: "NEUTRAL", details: ["Insufficient data"] };

  const price = closes[closes.length - 1],
    prev = closes[closes.length - 2];
  const upper = bb.upper[bb.upper.length - 1],
    middle = bb.middle[bb.middle.length - 1],
    lower = bb.lower[bb.lower.length - 1];
  const percentB = (price - lower) / (upper - lower);

  let signal = 0;
  const details = [];
  let position;
  if (price >= upper) {
    position = "ABOVE_UPPER";
    signal -= 25;
    details.push("At upper band");
  } else if (price <= lower) {
    position = "BELOW_LOWER";
    signal += 25;
    details.push("At lower band");
  } else if (price > middle) {
    position = "UPPER_HALF";
    signal += 15;
    details.push("Upper half");
  } else {
    position = "LOWER_HALF";
    signal -= 15;
    details.push("Lower half");
  }

  if (prev <= bb.lower[bb.lower.length - 2] && price > lower) {
    signal += 30;
    details.push("Bounced from lower");
  } else if (prev >= bb.upper[bb.upper.length - 2] && price < upper) {
    signal -= 30;
    details.push("Rejected from upper");
  }

  return {
    signal: Math.min(100, Math.max(-100, signal)),
    position,
    percentB,
    values: { upper, middle, lower },
    details,
  };
}
