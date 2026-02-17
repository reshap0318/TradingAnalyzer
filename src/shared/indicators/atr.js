import config from "../../config.js";

export function calculateATR(ohlcData, period = 14) {
  if (ohlcData.length < period + 1) return [];
  const tr = [];
  for (let i = 1; i < ohlcData.length; i++) {
    tr.push(
      Math.max(
        ohlcData[i].high - ohlcData[i].low,
        Math.abs(ohlcData[i].high - ohlcData[i - 1].close),
        Math.abs(ohlcData[i].low - ohlcData[i - 1].close)
      )
    );
  }
  let atr = tr.slice(0, period).reduce((a, b) => a + b, 0) / period;
  const atrValues = [atr];
  for (let i = period; i < tr.length; i++) {
    atr = (atr * (period - 1) + tr[i]) / period;
    atrValues.push(atr);
  }
  return atrValues;
}

export function analyzeATR(ohlcData) {
  const atrValues = calculateATR(ohlcData, config.INDICATORS.ATR.PERIOD);
  if (atrValues.length < 20)
    return { atr: 0, volatility: "UNKNOWN", details: ["Insufficient data"] };

  const current = atrValues[atrValues.length - 1];
  const price = ohlcData[ohlcData.length - 1].close;
  const percent = (current / price) * 100;
  const avg = atrValues.slice(-20).reduce((a, b) => a + b, 0) / 20;
  const ratio = current / avg;

  let volatility;
  const details = [];
  if (ratio >= 1.5) {
    volatility = "HIGH";
    details.push(`ATR ${percent.toFixed(2)}% High`);
  } else if (ratio <= 0.7) {
    volatility = "LOW";
    details.push(`ATR ${percent.toFixed(2)}% Low`);
  } else {
    volatility = "NORMAL";
    details.push(`ATR ${percent.toFixed(2)}%`);
  }

  return {
    atr: current,
    atrPercent: percent,
    volatility,
    atrRatio: ratio,
    details,
  };
}
