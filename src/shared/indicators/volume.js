import config from "../../config.js";
import { calculateSMA } from "./movingAverages.js";

export function analyzeVolume(ohlcData) {
  if (ohlcData.length < config.INDICATORS.VOLUME.MA_PERIOD + 5)
    return { signal: 0, confirmation: false, details: ["Insufficient data"] };

  const volumes = ohlcData.map((d) => d.volume),
    closes = ohlcData.map((d) => d.close);
  const curr = volumes[volumes.length - 1],
    currClose = closes[closes.length - 1],
    prevClose = closes[closes.length - 2];
  const avgVol = calculateSMA(volumes, config.INDICATORS.VOLUME.MA_PERIOD);
  const avg = avgVol[avgVol.length - 1];
  const ratio = curr / avg;

  let signal = 0;
  const details = [];
  let confirmation = false;

  if (ratio >= 2.0) {
    confirmation = true;
    if (currClose > prevClose) {
      signal += 40;
      details.push("Very high volume up");
    } else {
      signal -= 40;
      details.push("Very high volume down");
    }
  } else if (ratio >= 1.5) {
    confirmation = true;
    if (currClose > prevClose) {
      signal += 25;
      details.push("High volume up");
    } else {
      signal -= 25;
      details.push("High volume down");
    }
  } else if (ratio >= 1.0) {
    if (currClose > prevClose) signal += 10;
    else signal -= 10;
    details.push("Normal volume");
  } else {
    details.push("Low volume");
  }

  return {
    signal: Math.min(100, Math.max(-100, signal)),
    confirmation,
    volumeRatio: ratio,
    currentVolume: curr,
    avgVolume: avg,
    details,
  };
}
