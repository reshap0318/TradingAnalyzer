import config from "../config.js";

/**
 * Enhanced Money Management with user input
 * @param {Object} portfolio - User's portfolio info
 * @param {number} portfolio.totalCapital - Total capital (e.g., 100000000)
 * @param {number} portfolio.maxLossPercent - Maximum loss tolerance in % (e.g., 2 = 2%)
 * @param {number} portfolio.currentPositions - Current open positions
 * @param {number} price - Current stock price
 * @param {number} slPrice - Stop loss price
 * @param {Object} tpsl - TP/SL data with tp percentage
 * @param {string} signal - BUY/SELL/WAIT
 * @param {number} trendStrength - Trend strength score (0-100)
 */
export function calculateMoneyManagement(
  portfolio,
  price,
  slPrice,
  tpsl,
  signal,
  trendStrength = 50,
  assetType = "STOCK"
) {
  const { totalCapital, maxLossPercent = 2, currentPositions = 0 } = portfolio;

  // Calculate validity of the trade setup
  const tpPercent = tpsl?.tp?.percent ? Math.abs(tpsl.tp.percent) : 0;
  const slPercent = slPrice ? Math.abs(((slPrice - price) / price) * 100) : 0;

  const validSignal =
    assetType === "CRYPTO"
      ? signal === "BUY" || signal === "SELL"
      : signal === "BUY"; // Stock = long-only

  let isValid = validSignal && slPrice && tpPercent > slPercent;

  // If no stop loss, we cannot calculate risk
  if (!slPrice) {
    return {
      isValid: false,
      reason: "Stop loss not defined",
      recommendation: null,
    };
  }

  const riskPerShare = Math.abs(price - slPrice);
  const riskPerSharePercent = (riskPerShare / price) * 100;

  // Calculate max loss amount based on user's tolerance
  const maxLossAmount = totalCapital * (maxLossPercent / 100);

  // Calculate max shares based on risk
  let maxSharesByRisk = maxLossAmount / riskPerShare;

  // For Stocks, we must floor to integer shares (actually lots handles this, but maxShares should strictly be int for stocks?)
  // Actually, for stocks logic below: floor(shares / 100).
  // For Crypto, we keep decimal.
  if (assetType === "STOCK") {
    maxSharesByRisk = Math.floor(maxSharesByRisk);
  }

  // Convert to lots
  // IDX (Stock): 1 Lot = 100 Shares
  // Crypto: 1 Lot = 1 Unit (or fractional)
  const lotSize = assetType === "CRYPTO" ? 1 : 100;

  let maxLotsRaw;
  if (assetType === "CRYPTO") {
    maxLotsRaw = maxSharesByRisk; // Keep fractional
  } else {
    maxLotsRaw = Math.floor(maxSharesByRisk / lotSize);
  }

  // Adjust lots based on trend strength (stronger trend = more confidence)
  let lotMultiplier = 1;
  if (trendStrength >= 80) lotMultiplier = 1.0; // Strong trend - full position
  else if (trendStrength >= 60) lotMultiplier = 0.75; // Good trend - 75%
  else if (trendStrength >= 40) lotMultiplier = 0.5; // Moderate - 50%
  else lotMultiplier = 0.25; // Weak - 25% only

  // Recommended lots
  let recommendedLots;
  if (assetType === "CRYPTO") {
    recommendedLots = parseFloat((maxLotsRaw * lotMultiplier).toFixed(8)); // 8 decimals for crypto
  } else {
    recommendedLots = Math.floor(maxLotsRaw * lotMultiplier);
    // Prevent lot multiplier from dropping a 1-lot capacity down to 0
    if (recommendedLots === 0 && maxLotsRaw >= 1) {
      recommendedLots = 1;
    }
  }

  const recommendedShares = recommendedLots * lotSize;

  const positionValue = recommendedShares * price;
  const actualRiskAmount = recommendedShares * riskPerShare;
  const actualRiskPercent = (actualRiskAmount / totalCapital) * 100;

  // Calculate potential profit at each TP
  const potentialProfit = {
    tp: tpsl?.tp ? recommendedShares * Math.abs(tpsl.tp.price - price) : 0,
  };

  // Risk/Reward ratio
  const riskRewardRatio = slPercent > 0 ? tpPercent / slPercent : 0;

  const warnings = [];
  if (actualRiskPercent > maxLossPercent * 0.9)
    warnings.push("Mendekati batas maksimal kerugian");
  if (positionValue > totalCapital * 0.1)
    warnings.push("Posisi > 10% dari total capital");
  if (riskRewardRatio < 1.5) warnings.push("Risk/Reward ratio kurang dari 1.5");
  if (!validSignal && signal !== "WAIT")
    warnings.push("Signal is not valid for this asset type");
  if (validSignal && slPrice && tpPercent <= slPercent)
    warnings.push(
      `TP (${tpPercent.toFixed(2)}%) < SL (${slPercent.toFixed(2)}%)`
    );

  const finalPositionValue =
    assetType === "CRYPTO"
      ? Number(positionValue.toFixed(4)) // Preserve decimals to prevent rounding down to 0 for small crypto allocations
      : Math.round(positionValue);

  return {
    isValid,
    signal,
    totalLot: recommendedLots,
    priceLot: {
      satuan: price,
      totalBelanja: finalPositionValue,
    },
    tpPercent: Math.round(tpPercent * 100) / 100,
    slPercent: Math.round(slPercent * 100) / 100,
    riskRewardRatio: Math.round(riskRewardRatio * 100) / 100,
    maksimalKerugian:
      assetType === "CRYPTO"
        ? Math.round(actualRiskAmount * 100) / 100
        : Math.round(actualRiskAmount),
    potensiKeuntungan:
      assetType === "CRYPTO"
        ? Math.round(potentialProfit.tp * 100) / 100
        : Math.round(potentialProfit.tp),
    warnings,
  };
}

/**
 * Check if trade should be taken based on TP/SL ratio
 */
export function shouldTrade(tpsl, minProfitLossRatio = 1.5) {
  if (!tpsl || !tpsl.tp || !tpsl.sl)
    return { trade: false, reason: "TP atau SL tidak tersedia" };

  const tpPercent = Math.abs(tpsl.tp.percent);
  const slPercent = Math.abs(tpsl.sl.percent);
  const ratio = tpPercent / slPercent;

  if (ratio < minProfitLossRatio) {
    return {
      trade: false,
      reason: `Risk/Reward ratio ${ratio.toFixed(
        2
      )} kurang dari minimum ${minProfitLossRatio}`,
      ratio,
    };
  }

  return { trade: true, ratio };
}
