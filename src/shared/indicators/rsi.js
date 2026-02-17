import config from "../../config.js";

export function calculateRSI(closes, period = 14) {
  if (closes.length < period + 1) return [];
  const gains = [],
    losses = [];
  for (let i = 1; i < closes.length; i++) {
    const change = closes[i] - closes[i - 1];
    gains.push(change > 0 ? change : 0);
    losses.push(change < 0 ? Math.abs(change) : 0);
  }
  let avgGain = gains.slice(0, period).reduce((a, b) => a + b, 0) / period;
  let avgLoss = losses.slice(0, period).reduce((a, b) => a + b, 0) / period;
  const rsi = [100 - 100 / (1 + (avgLoss === 0 ? 100 : avgGain / avgLoss))];
  for (let i = period; i < gains.length; i++) {
    avgGain = (avgGain * (period - 1) + gains[i]) / period;
    avgLoss = (avgLoss * (period - 1) + losses[i]) / period;
    rsi.push(100 - 100 / (1 + (avgLoss === 0 ? 100 : avgGain / avgLoss)));
  }
  return rsi;
}

export function analyzeRSI(closes) {
  const rsi = calculateRSI(closes, config.INDICATORS.RSI.PERIOD);
  if (rsi.length < 5)
    return {
      signal: 0,
      value: 0,
      zone: "NEUTRAL",
      details: ["Insufficient data"],
    };

  const current = rsi[rsi.length - 1],
    prev = rsi[rsi.length - 2];
  let signal = 0;
  const details = [];
  let zone;

  if (current >= 70) {
    zone = "OVERBOUGHT";
    signal -= 30;
    details.push(`RSI ${current.toFixed(1)} Overbought`);
  } else if (current <= 30) {
    zone = "OVERSOLD";
    signal += 30;
    details.push(`RSI ${current.toFixed(1)} Oversold`);
  } else if (current > 50) {
    zone = "BULLISH";
    signal += 20;
    details.push(`RSI ${current.toFixed(1)} Bullish`);
  } else {
    zone = "BEARISH";
    signal -= 20;
    details.push(`RSI ${current.toFixed(1)} Bearish`);
  }

  if (prev < 50 && current >= 50) {
    signal += 20;
    details.push("RSI crossed above 50");
  } else if (prev > 50 && current <= 50) {
    signal -= 20;
    details.push("RSI crossed below 50");
  }

  return {
    signal: Math.min(100, Math.max(-100, signal)),
    value: current,
    zone,
    details,
  };
}
