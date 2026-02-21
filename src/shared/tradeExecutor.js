import config from "../config.js";
import {
  placeFuturesOrder,
  setTakeProfitStopLoss,
} from "../crypto/binanceTrader.js";
import { getSummary } from "./signalLogger.js";

let tradesToday = 0;
let lastReset = new Date().getDate();

// Reset trades count every day at midnight
function resetDailyTrades() {
  const currentDay = new Date().getDate();
  if (currentDay !== lastReset) {
    tradesToday = 0;
    lastReset = currentDay;
  }
}

export function getTradingStatus(assetType = "CRYPTO") {
  const summary = getSummary(assetType);
  const initialCap =
    assetType === "CRYPTO"
      ? config.CRYPTO.DEFAULT_CAPITAL
      : config.SAHAM.DEFAULT_CAPITAL;
  const drawdownAmount = -Math.min(0, summary.totalPnlDollar || 0);
  const drawdownPercent = (drawdownAmount / initialCap) * 100;

  return {
    enabled: config.AUTO_TRADING.ENABLED,
    active:
      config.AUTO_TRADING.ENABLED &&
      tradesToday < config.AUTO_TRADING.SAFETY.MAX_DAILY_TRADES &&
      drawdownPercent < config.AUTO_TRADING.SAFETY.MAX_DRAWDOWN_PERCENT,
    network: config.AUTO_TRADING.USE_TESTNET ? "testnet" : "prod",
    todayTrades: tradesToday,
    drawdownPercent: Math.round(drawdownPercent * 100) / 100,
    maxDailyTradesTriggered:
      tradesToday >= config.AUTO_TRADING.SAFETY.MAX_DAILY_TRADES,
    maxDrawdownTriggered:
      drawdownPercent >= config.AUTO_TRADING.SAFETY.MAX_DRAWDOWN_PERCENT,
  };
}

export function checkSafetyRules(
  analysis,
  assetType = "CRYPTO",
  isSimulation = false
) {
  resetDailyTrades();
  const status = getTradingStatus(assetType);
  const reasons = [];

  if (!isSimulation) {
    if (!status.enabled)
      reasons.push("Auto-Trading is globally disabled in config.js");
    if (!status.active)
      reasons.push("Bot is in Standby (Maximum limits reached)");
    if (status.maxDailyTradesTriggered)
      reasons.push(
        `Daily Trade Limit Hit (${config.AUTO_TRADING.SAFETY.MAX_DAILY_TRADES})`
      );
    if (status.maxDrawdownTriggered)
      reasons.push(
        `Maximum Drawdown Limit Hit (${config.AUTO_TRADING.SAFETY.MAX_DRAWDOWN_PERCENT}%)`
      );
  }

  if (analysis.signalResult?.signal === "WAIT") reasons.push("Signal is WAIT");
  if (
    (analysis.signalResult?.confidence || 0) <
    config.AUTO_TRADING.SAFETY.MIN_CONFIDENCE
  )
    reasons.push(
      `Confidence (${analysis.signalResult?.confidence}%) < Target (${config.AUTO_TRADING.SAFETY.MIN_CONFIDENCE}%)`
    );
  // Only block if explicitly set to invalid. If money Mgmt isn't passed at all, leave it.
  if (analysis.moneyMgmt !== undefined && analysis.moneyMgmt?.isValid === false)
    reasons.push(
      `Money Management rejected trade: ${
        analysis.moneyMgmt?.warnings?.join(", ") || "Invalid Risk"
      }`
    );

  // Checking RR
  if (
    analysis.tpsl?.riskReward?.tp &&
    analysis.tpsl.riskReward.tp < config.AUTO_TRADING.SAFETY.MIN_RISK_REWARD
  ) {
    reasons.push(
      `Risk/Reward (${analysis.tpsl.riskReward.tp}) < Minimum (${config.AUTO_TRADING.SAFETY.MIN_RISK_REWARD})`
    );
  }

  // Crash checks via market sentiment
  if (config.AUTO_TRADING.SAFETY.BLOCK_ON_CRASH) {
    if (assetType === "CRYPTO" && analysis.btcMarket?.isCrash) {
      reasons.push("Market is currently marked as CRASH");
    } else if (assetType === "SAHAM" && analysis.ihsgAnalysis?.isCrash) {
      reasons.push("IHSG is currently in a CRASH state");
    }
  }

  return { safe: reasons.length === 0, reasons };
}

export async function executeTrade(analysis) {
  const safety = checkSafetyRules(analysis);

  if (!safety.safe) {
    console.log(
      `[TRADE_BLOCKED] ${analysis.quote?.symbol} -> ${safety.reasons[0]}`
    );
    return { executed: false, reasons: safety.reasons };
  }

  const symbol = analysis.quote.symbol;
  const side = analysis.signalResult.signal;
  const quantity =
    analysis.moneyMgmt?.totalLot || analysis.moneyMgmt?.alokasiDana?.lot;
  const leverage = config.AUTO_TRADING.DEFAULT_LEVERAGE;

  if (!quantity || quantity <= 0) {
    return {
      executed: false,
      reasons: ["Calculated Lot quantity is zero or invalid"],
    };
  }

  try {
    tradesToday++;
    // 1. Enter the Position
    const order = await placeFuturesOrder({
      symbol,
      side,
      quantity,
      leverage,
      type: config.AUTO_TRADING.ORDER_TYPE,
    });

    let tpSlOrders = null;

    // 2. Set TP/SL Bounds
    if (
      config.AUTO_TRADING.AUTO_SET_TP_SL &&
      analysis.tpsl?.tp?.price &&
      analysis.tpsl?.sl?.price
    ) {
      tpSlOrders = await setTakeProfitStopLoss({
        symbol,
        side,
        tpPrice: analysis.tpsl.tp.price,
        slPrice: analysis.tpsl.sl.price,
        quantity,
        parentOrderId: order.orderId,
      });
    }

    return {
      executed: true,
      network: config.AUTO_TRADING.USE_TESTNET ? "testnet" : "prod",
      order,
      tpSlOrders,
    };
  } catch (error) {
    console.error(`💥 Execution fault on ${symbol}:`, error.message);
    return { executed: false, error: error.message };
  }
}
