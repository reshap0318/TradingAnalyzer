import config from "../config.js";

/**
 * Futures Calculator — Crypto-specific
 * Calculates leverage-adjusted position sizing, liquidation prices,
 * margin requirements, and funding cost estimates.
 *
 * This module is ONLY for crypto futures, never used by saham.
 */

const futuresConfig = () => config.CRYPTO?.FUTURES || {};

/**
 * Calculate full futures trade plan
 *
 * @param {Object} params
 * @param {number} params.capital - Total capital in USD
 * @param {number} params.entryPrice - Entry price
 * @param {number} params.slPrice - Stop loss price
 * @param {string} params.side - "LONG" or "SHORT"
 * @param {number} params.leverage - Leverage multiplier (e.g., 5)
 * @param {Object} params.tpsl - TP/SL object from tpslCalculator
 * @returns {Object} Futures trade plan
 */
export function calculateFuturesPlan({
  capital,
  entryPrice,
  slPrice,
  side = "LONG",
  leverage = null,
  tpsl = {},
}) {
  const cfg = futuresConfig();
  const lev = Math.min(
    leverage || cfg.DEFAULT_LEVERAGE || 5,
    cfg.MAX_LEVERAGE || 20
  );
  const maintenanceRate = cfg.MAINTENANCE_MARGIN_RATE || 0.005;
  const takerFee = cfg.TAKER_FEE || 0.0004;
  const maxPositionPercent = cfg.MAX_POSITION_PERCENT || 0.1;

  const isLong = side === "LONG";

  // --- Margin & Position Sizing ---
  const maxPositionValue = capital * maxPositionPercent;
  const margin = maxPositionValue; // Margin = capital allocated to this trade
  const positionValue = margin * lev; // Notional value with leverage
  const quantity = positionValue / entryPrice;

  // --- Liquidation Price ---
  // Long:  liqPrice = entryPrice * (1 - 1/leverage + maintenanceRate)
  // Short: liqPrice = entryPrice * (1 + 1/leverage - maintenanceRate)
  const liquidationPrice = isLong
    ? entryPrice * (1 - 1 / lev + maintenanceRate)
    : entryPrice * (1 + 1 / lev - maintenanceRate);

  // --- Risk Calculation ---
  const slDistance = Math.abs(entryPrice - slPrice);
  const slPercent = (slDistance / entryPrice) * 100;
  const slLoss = quantity * slDistance; // Actual USD loss if SL hit
  const slLossPercent = (slLoss / capital) * 100; // Loss as % of total capital

  // --- Check if SL is beyond liquidation ---
  const slBeyondLiq = isLong
    ? slPrice <= liquidationPrice
    : slPrice >= liquidationPrice;

  // --- TP Calculations with leverage ---
  const tpCalc = (tpPrice) => {
    if (!tpPrice) return null;
    const diff = isLong ? tpPrice - entryPrice : entryPrice - tpPrice;
    const pnl = quantity * diff;
    const roe = (diff / entryPrice) * lev * 100; // Return on Equity
    return {
      price: tpPrice,
      pnl: Math.round(pnl * 100) / 100,
      roe: Math.round(roe * 100) / 100,
    };
  };

  // --- Fee Estimation ---
  const openFee = positionValue * takerFee;
  const closeFee = positionValue * takerFee;
  const totalFees = openFee + closeFee;

  // --- Funding Rate Estimate (per 8h) ---
  const fundingInterval = cfg.FUNDING_INTERVAL_HOURS || 8;
  const estimatedFundingRate = 0.0001; // 0.01% per interval (typical)
  const fundingPer8h = positionValue * estimatedFundingRate;

  // --- Effective PnL at each TP (minus fees) ---
  const tpData = tpCalc(tpsl.tp?.price);

  // --- Warnings ---
  const warnings = [];
  if (slBeyondLiq) {
    warnings.push(
      `⚠️ Stop loss (${slPrice}) melewati liquidation price (${
        Math.round(liquidationPrice * 100) / 100
      }). Posisi akan di-liquidate sebelum SL!`
    );
  }
  if (slLossPercent > 5) {
    warnings.push(
      `Kerugian jika SL kena = ${slLossPercent.toFixed(
        1
      )}% dari capital. Pertimbangkan leverage lebih rendah.`
    );
  }
  if (lev > 10) {
    warnings.push(`Leverage ${lev}x sangat tinggi. Risiko liquidasi besar.`);
  }
  if (totalFees > margin * 0.02) {
    warnings.push(
      `Fee trading (${totalFees.toFixed(2)} USD) > 2% dari margin.`
    );
  }

  return {
    side,
    leverage: lev,
    margin: {
      required: Math.round(margin * 100) / 100,
      maintenanceRate: maintenanceRate * 100 + "%",
    },
    position: {
      notionalValue: Math.round(positionValue * 100) / 100,
      quantity: Math.round(quantity * 100000000) / 100000000,
      entryPrice,
    },
    liquidation: {
      price: Math.round(liquidationPrice * 100) / 100,
      distancePercent:
        Math.round(
          (Math.abs(entryPrice - liquidationPrice) / entryPrice) * 10000
        ) / 100,
      slBeyondLiquidation: slBeyondLiq,
    },
    risk: {
      slPrice,
      slPercent: Math.round(slPercent * 100) / 100,
      slLoss: Math.round(slLoss * 100) / 100,
      slLossOfCapital: Math.round(slLossPercent * 100) / 100 + "%",
      effectiveRisk: Math.round(slPercent * lev * 100) / 100 + "% ROE",
    },
    targets: {
      tp: tpData,
    },
    fees: {
      openFee: Math.round(openFee * 100) / 100,
      closeFee: Math.round(closeFee * 100) / 100,
      totalFees: Math.round(totalFees * 100) / 100,
      fundingPer8h: Math.round(fundingPer8h * 100) / 100,
    },
    binance_execution: {
      side: side,
      leverage: lev,
      marginMode: "ISOLATED",
      entryPrice: entryPrice,
      marginUsd: Math.round(margin * 100) / 100,
      notionalUsd: Math.round(positionValue * 100) / 100,
      takeProfit: tpData?.price || null,
      stopLoss: slPrice || null,
    },
    warnings,
  };
}
