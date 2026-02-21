import config, { TIMEFRAME_MAP } from "../config.js";
import {
  fetchMultiTimeframe,
  fetchIHSG,
  getCurrentPrice,
} from "./yahooFinance.js";
import { analyzeIHSG, calculateCorrelation } from "./ihsgAnalyzer.js";
import { makeDecision } from "./decisionEngine.js";
import { generateSignal } from "../shared/signalGenerator.js";
import { calculateTPSL } from "../shared/tpslCalculator.js";
import { calculateMoneyManagement } from "../shared/moneyManagement.js";
import { getCapitalStatus } from "../shared/signalLogger.js";

export async function analyzeSahamSymbol(
  symbol,
  interval,
  initialCapital = config.SAHAM.DEFAULT_CAPITAL,
  maxLossPercent = null
) {
  const timeframes = TIMEFRAME_MAP[interval] || TIMEFRAME_MAP["1d"];
  const capitalStatus = getCapitalStatus("SAHAM", initialCapital);
  const portfolio = {
    totalCapital: capitalStatus.available,
    maxLossPercent:
      maxLossPercent || config.SAHAM.MONEY_MANAGEMENT.MAX_RISK_PER_TRADE * 100,
    currentPositions: capitalStatus.openPositions,
  };

  const [multiTfData, ihsgData, quote] = await Promise.all([
    fetchMultiTimeframe(symbol, timeframes),
    fetchIHSG(),
    getCurrentPrice(symbol),
  ]);
  const primaryData = multiTfData[interval];
  if (!primaryData || primaryData.length < 50)
    throw new Error("Insufficient data for " + interval);

  const ihsgAnalysis = analyzeIHSG(ihsgData);
  const correlation = calculateCorrelation(primaryData, ihsgData);
  const decision = makeDecision(
    primaryData,
    multiTfData,
    ihsgAnalysis,
    timeframes
  );
  const signalResult = generateSignal(decision, quote.price, symbol);
  const tpsl = calculateTPSL(primaryData, signalResult.signal, quote.price);

  const trendStrength = Math.abs(decision.score);

  const moneyMgmt = calculateMoneyManagement(
    portfolio,
    quote.price,
    tpsl.sl?.price,
    tpsl,
    signalResult.signal,
    trendStrength,
    "SAHAM"
  );

  return {
    quote,
    capitalStatus,
    ihsgAnalysis,
    correlation,
    decision,
    signalResult,
    tpsl,
    moneyMgmt,
    timeframes,
  };
}
