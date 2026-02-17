import config from "../config.js";

/**
 * Find Pivot Highs (LuxAlgo Style)
 * Checks left and right bars to confirm local maximum
 */
function getPivotHighs(highs, leftBars, rightBars) {
  let pivots = [];
  for (let i = leftBars; i < highs.length - rightBars; i++) {
    let isPivot = true;
    let currentHigh = highs[i];

    // Check left and right neighbors
    for (let j = i - leftBars; j <= i + rightBars; j++) {
      if (i !== j && highs[j] >= currentHigh) {
        isPivot = false;
        break;
      }
    }
    if (isPivot) pivots.push(currentHigh);
  }
  return pivots;
}

/**
 * Find Pivot Lows (LuxAlgo Style)
 * Checks left and right bars to confirm local minimum
 */
function getPivotLows(lows, leftBars, rightBars) {
  let pivots = [];
  for (let i = leftBars; i < lows.length - rightBars; i++) {
    let isPivot = true;
    let currentLow = lows[i];

    // Check left and right neighbors
    for (let j = i - leftBars; j <= i + rightBars; j++) {
      if (i !== j && lows[j] <= currentLow) {
        isPivot = false;
        break;
      }
    }
    if (isPivot) pivots.push(currentLow);
  }
  return pivots;
}

/**
 * Find Support and Resistance Levels using LuxAlgo Logic
 * @param {Array} ohlcData - Array of OHLC objects
 * @returns {Object} - Supports and Resistances + Raw Pivots
 */
export function findSupportResistance(ohlcData) {
  const { LOOKBACK_PERIODS } = config.SR_SETTINGS; // Use lookback for slicing mostly

  if (!ohlcData || ohlcData.length < 50) {
    return { supports: [], resistances: [], pivotHighs: [], pivotLows: [] };
  }

  // Use mostly recent data but enough for pivots
  // LuxAlgo default: left 15, right 15
  const leftBars = 15;
  const rightBars = 15;

  // We need enough history to find pivots.
  // If we slice too small, we miss old pivots.
  // Let's use last 300 candles if available
  const data = ohlcData.slice(-300);

  const highs = data.map((d) => d.high);
  const lows = data.map((d) => d.low);
  const currentPrice = data[data.length - 1].close;

  const pivotHighsList = getPivotHighs(highs, leftBars, rightBars);
  const pivotLowsList = getPivotLows(lows, leftBars, rightBars);

  // Filter and sort for nearest S/R
  let resistances = [...new Set(pivotHighsList)]
    .filter((r) => r > currentPrice)
    .sort((a, b) => a - b) // Nearest up
    .map((level) => ({ level, type: "PIVOT" }))
    .slice(0, 3);

  let supports = [...new Set(pivotLowsList)]
    .filter((s) => s < currentPrice)
    .sort((a, b) => b - a) // Nearest down
    .map((level) => ({ level, type: "PIVOT" }))
    .slice(0, 3);

  // Fallback: If no resistance found (ATH), use Max High
  if (resistances.length === 0) {
    const maxHigh = Math.max(...highs.slice(-20));
    if (maxHigh > currentPrice) {
      resistances.push({ level: maxHigh, type: "ATH/MAX" });
    }
  }

  // Fallback: If no support found (ATL), use Min Low
  if (supports.length === 0) {
    const minLow = Math.min(...lows.slice(-20));
    if (minLow < currentPrice) {
      supports.push({ level: minLow, type: "ATL/MIN" });
    }
  }

  return {
    currentPrice,
    supports,
    resistances,
    pivotHighs: pivotHighsList,
    pivotLows: pivotLowsList,
  };
}
