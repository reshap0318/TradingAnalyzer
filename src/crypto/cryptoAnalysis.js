import config, { TIMEFRAME_MAP } from "../config.js";
import {
  fetchMultiTimeframe as fetchBinanceMultiTf,
  fetchBTCDominance,
  getCurrentPrice as getBinancePrice,
} from "./binanceData.js";
import { analyzeBTCMarket } from "./btcMarketAnalyzer.js";
import { makeCryptoDecision } from "./cryptoDecisionEngine.js";
import { generateSignal } from "../shared/signalGenerator.js";
import { calculateTPSL } from "../shared/tpslCalculator.js";
import { calculateMoneyManagement } from "../shared/moneyManagement.js";
import { getCapitalStatus } from "../shared/signalLogger.js";

export async function analyzeCryptoSymbol(
  symbol,
  interval,
  leverage,
  initialCapital = config.CRYPTO.DEFAULT_CAPITAL,
  maxLossPercent = null
) {
  const timeframes = TIMEFRAME_MAP[interval] || TIMEFRAME_MAP["1d"];
  const capitalStatus = getCapitalStatus("CRYPTO", initialCapital);
  const portfolio = {
    totalCapital: capitalStatus.available,
    maxLossPercent:
      maxLossPercent || config.CRYPTO.MONEY_MANAGEMENT.MAX_RISK_PER_TRADE * 100,
    currentPositions: capitalStatus.openPositions,
  };

  const [multiTfData, btcDomData, quote] = await Promise.all([
    fetchBinanceMultiTf(symbol, timeframes),
    fetchBTCDominance(),
    getBinancePrice(symbol),
  ]);

  const primaryData = multiTfData[interval];
  if (!primaryData || primaryData.length < 30)
    throw new Error("Insufficient data for " + interval);

  const btcMarket = analyzeBTCMarket(btcDomData, symbol);
  const decision = makeCryptoDecision(
    primaryData,
    multiTfData,
    btcMarket,
    timeframes
  );

  const signalResult = generateSignal(
    decision,
    quote.price,
    symbol,
    "BTC Market"
  );

  const tpsl = calculateTPSL(
    primaryData,
    signalResult.signal,
    quote.price,
    "CRYPTO"
  );

  const trendStrength = Math.abs(decision.score);

  const moneyMgmt = calculateMoneyManagement(
    portfolio,
    quote.price,
    tpsl.sl?.price,
    tpsl,
    signalResult.signal,
    trendStrength,
    "CRYPTO"
  );

  return {
    symbol,
    interval,
    quote,
    capitalStatus,
    btcMarket,
    decision,
    signalResult,
    tpsl,
    moneyMgmt,
    leverage,
  };
}
