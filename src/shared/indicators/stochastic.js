import config from "../../config.js";

export function calculateStochastic(highs, lows, closes) {
  const { K_PERIOD, D_PERIOD, SMOOTH } = config.INDICATORS.STOCHASTIC;
  if (closes.length < K_PERIOD + D_PERIOD) return { k: [], d: [] };

  const rawK = [];
  for (let i = K_PERIOD - 1; i < closes.length; i++) {
    const hh = Math.max(...highs.slice(i - K_PERIOD + 1, i + 1));
    const ll = Math.min(...lows.slice(i - K_PERIOD + 1, i + 1));
    rawK.push(hh === ll ? 50 : ((closes[i] - ll) / (hh - ll)) * 100);
  }

  const k = [];
  for (let i = SMOOTH - 1; i < rawK.length; i++)
    k.push(
      rawK.slice(i - SMOOTH + 1, i + 1).reduce((a, b) => a + b, 0) / SMOOTH
    );
  const d = [];
  for (let i = D_PERIOD - 1; i < k.length; i++)
    d.push(
      k.slice(i - D_PERIOD + 1, i + 1).reduce((a, b) => a + b, 0) / D_PERIOD
    );
  return { k, d };
}

export function analyzeStochastic(ohlcData) {
  const stoch = calculateStochastic(
    ohlcData.map((d) => d.high),
    ohlcData.map((d) => d.low),
    ohlcData.map((d) => d.close)
  );
  if (stoch.k.length < 3 || stoch.d.length < 3)
    return { signal: 0, zone: "NEUTRAL", details: ["Insufficient data"] };

  const currK = stoch.k[stoch.k.length - 1],
    currD = stoch.d[stoch.d.length - 1];
  const prevK = stoch.k[stoch.k.length - 2],
    prevD = stoch.d[stoch.d.length - 2];

  let signal = 0;
  const details = [];
  let zone;
  if (currK >= 80) {
    zone = "OVERBOUGHT";
    signal -= 20;
    details.push(`%K ${currK.toFixed(1)} Overbought`);
  } else if (currK <= 20) {
    zone = "OVERSOLD";
    signal += 20;
    details.push(`%K ${currK.toFixed(1)} Oversold`);
  } else {
    zone = "NEUTRAL";
    details.push(`%K ${currK.toFixed(1)}`);
  }

  if (prevK <= prevD && currK > currD) {
    signal += 35;
    details.push("%K crossed above %D");
  } else if (prevK >= prevD && currK < currD) {
    signal -= 35;
    details.push("%K crossed below %D");
  } else if (currK > currD) {
    signal += 15;
    details.push("%K above %D");
  } else {
    signal -= 15;
    details.push("%K below %D");
  }

  return {
    signal: Math.min(100, Math.max(-100, signal)),
    zone,
    values: { k: currK, d: currD },
    details,
  };
}
