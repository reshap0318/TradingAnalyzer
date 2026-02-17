import config from "../../config.js";

export function calculateSMA(data, period) {
  const result = [];
  for (let i = period - 1; i < data.length; i++) {
    result.push(
      data.slice(i - period + 1, i + 1).reduce((a, b) => a + b, 0) / period
    );
  }
  return result;
}

export function calculateEMA(data, period) {
  const result = [];
  const multiplier = 2 / (period + 1);
  let ema = data.slice(0, period).reduce((a, b) => a + b, 0) / period;
  result.push(ema);
  for (let i = period; i < data.length; i++) {
    ema = (data[i] - ema) * multiplier + ema;
    result.push(ema);
  }
  return result;
}

export function analyzeMA(closes) {
  if (closes.length < 200)
    return { signal: 0, trend: "NEUTRAL", details: ["Insufficient data"] };

  const curr = closes[closes.length - 1];
  const sma20 = calculateSMA(closes, 20),
    sma50 = calculateSMA(closes, 50),
    sma200 = calculateSMA(closes, 200);
  const ema12 = calculateEMA(closes, 12),
    ema26 = calculateEMA(closes, 26);
  const lastSma20 = sma20[sma20.length - 1],
    lastSma50 = sma50[sma50.length - 1],
    lastSma200 = sma200[sma200.length - 1];
  const lastEma12 = ema12[ema12.length - 1],
    lastEma26 = ema26[ema26.length - 1];

  let signal = 0;
  const details = [];
  if (lastEma12 > lastEma26) {
    signal += 25;
    details.push("EMA12 > EMA26");
  } else {
    signal -= 25;
    details.push("EMA12 < EMA26");
  }
  if (lastSma20 > lastSma50 && lastSma50 > lastSma200) {
    signal += 35;
    details.push("SMA bullish");
  } else if (lastSma20 < lastSma50 && lastSma50 < lastSma200) {
    signal -= 35;
    details.push("SMA bearish");
  }
  if (curr > lastSma200) {
    signal += 20;
    details.push("Above SMA200");
  } else {
    signal -= 20;
    details.push("Below SMA200");
  }
  if (curr > lastSma20) {
    signal += 20;
    details.push("Above SMA20");
  } else {
    signal -= 20;
    details.push("Below SMA20");
  }

  return {
    signal: Math.min(100, Math.max(-100, signal)),
    trend: signal >= 50 ? "BULLISH" : signal <= -50 ? "BEARISH" : "NEUTRAL",
    values: { sma20: lastSma20, sma50: lastSma50, sma200: lastSma200 },
    details,
  };
}
