import config from "../../config.js";
import { calculateEMA } from "./movingAverages.js";

export function calculateMACD(closes) {
  const { FAST, SLOW, SIGNAL } = config.INDICATORS.MACD;
  if (closes.length < SLOW + SIGNAL)
    return { macdLine: [], signalLine: [], histogram: [] };

  const emaFast = calculateEMA(closes, FAST),
    emaSlow = calculateEMA(closes, SLOW);
  const macdLine = [];
  for (let i = 0; i < emaSlow.length; i++)
    macdLine.push(emaFast[i + SLOW - FAST] - emaSlow[i]);
  const signalLine = calculateEMA(macdLine, SIGNAL);
  const histogram = [];
  for (let i = 0; i < signalLine.length; i++)
    histogram.push(macdLine[i + SIGNAL - 1] - signalLine[i]);
  return { macdLine, signalLine, histogram };
}

export function analyzeMACD(closes) {
  const { macdLine, signalLine, histogram } = calculateMACD(closes);
  if (histogram.length < 3)
    return { signal: 0, details: ["Insufficient data"] };

  const currMACD = macdLine[macdLine.length - 1],
    currSig = signalLine[signalLine.length - 1];
  const prevMACD = macdLine[macdLine.length - 2],
    prevSig = signalLine[signalLine.length - 2];
  const currHist = histogram[histogram.length - 1],
    prevHist = histogram[histogram.length - 2];

  let signal = 0;
  const details = [];
  if (prevMACD <= prevSig && currMACD > currSig) {
    signal += 40;
    details.push("MACD bullish cross");
  } else if (prevMACD >= prevSig && currMACD < currSig) {
    signal -= 40;
    details.push("MACD bearish cross");
  } else if (currMACD > currSig) {
    signal += 20;
    details.push("MACD above signal");
  } else {
    signal -= 20;
    details.push("MACD below signal");
  }

  if (currMACD > 0) {
    signal += 15;
    details.push("MACD above zero");
  } else {
    signal -= 15;
    details.push("MACD below zero");
  }
  if (currHist > prevHist) {
    signal += 15;
    details.push("Histogram rising");
  } else {
    signal -= 15;
    details.push("Histogram falling");
  }

  return {
    signal: Math.min(100, Math.max(-100, signal)),
    values: { macd: currMACD, signal: currSig, histogram: currHist },
    details,
  };
}
