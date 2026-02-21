import config, { TIMEFRAME_MAP } from "../config.js";
import { analyzeSahamSymbol } from "./sahamAnalysis.js";
import { checkSafetyRules } from "../shared/tradeExecutor.js";
import { analyzeIHSG } from "./ihsgAnalyzer.js";
import { makeDecision } from "./decisionEngine.js";
import { generateSignal } from "../shared/signalGenerator.js";
import { calculateTPSL } from "../shared/tpslCalculator.js";
import {
  fetchMultiTimeframe,
  fetchIHSG,
  getCurrentPrice,
} from "./yahooFinance.js";
import {
  logSignal,
  closePendingSignal,
  updateOutcomes,
  getSummary as getSignalSummary,
  getHistory as getSignalHistory,
  getCapitalStatus,
} from "../shared/signalLogger.js";

const getCurrentWIB = () => {
  return new Date().toLocaleString("id-ID", {
    timeZone: "Asia/Jakarta",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
};

const formatSymbol = (s) =>
  s.includes(".") ? s.toUpperCase() : `${s.toUpperCase()}.JK`;

export const analyzeSaham = async (req, res) => {
  try {
    if (!req.query.symbol) {
      return res.status(400).json({ error: "Symbol parameter is required" });
    }
    const symbol = formatSymbol(req.query.symbol);
    const interval = req.query.interval || config.DEFAULT_INTERVAL;
    const initialCapital =
      parseInt(req.query.capital) || config.SAHAM.DEFAULT_CAPITAL;
    const maxLossInput = req.query.maxLoss || req.query.maxloss;
    const maxLossPercent = maxLossInput ? parseFloat(maxLossInput) : null;

    console.log(`\n📊 Analyzing ${symbol} [${interval}]...`);

    const analysis = await analyzeSahamSymbol(
      symbol,
      interval,
      initialCapital,
      maxLossPercent
    );
    const {
      quote,
      capitalStatus,
      ihsgAnalysis,
      correlation,
      decision,
      signalResult,
      tpsl,
      moneyMgmt,
    } = analysis;

    console.log(`✅ ${signalResult.signal} (${signalResult.confidence}%)`);

    // Update outcomes of pending saham signals
    updateOutcomes(async (sym) => {
      try {
        const q = await getCurrentPrice(sym);
        return q.price;
      } catch {
        return null;
      }
    }, "SAHAM").catch(() => {});

    res.json({
      symbol,
      interval,
      timestamp: getCurrentWIB(),
      capitalStatus,
      currentPrice: quote.price,
      change: quote.change,
      changePercent: quote.changePercent,
      volume: quote.volume,
      trade_plan: {
        valid: signalResult.signal === "BUY",
        signal: signalResult.signal,
        strength: signalResult.strength,
        entry: signalResult.entryZone,
        tp: tpsl.tp,
        sl: tpsl.sl,
        riskReward: tpsl.riskReward,
      },
      scoring: {
        totalScore: signalResult.score,
        confidence: signalResult.confidence,
        breakdown: decision.breakdown || [],
      },
      market_sentiment: {
        index: "IHSG",
        trend: ihsgAnalysis.trend,
        strength: ihsgAnalysis.strength,
        change1d: ihsgAnalysis.change1d,
        change5d: ihsgAnalysis.change5d,
        isCrash: ihsgAnalysis.isCrash,
        correlation: Math.round(correlation * 100) / 100,
        details: ihsgAnalysis.details,
      },
      timeframes: Object.fromEntries(
        Object.entries(decision.multiTimeframe.timeframes).map(([tf, d]) => [
          tf,
          { trend: d.trend, signal: Math.round(d.signal) },
        ])
      ),
      timeframeAlignment: decision.multiTimeframe.alignment,
      reasoning: signalResult.reasoning,
      warnings: signalResult.warnings,
      moneyManagement: moneyMgmt,
      patterns: decision.patterns,
      indicators: {
        ma: decision.indicators.ma,
        rsi: decision.indicators.rsi,
        macd: decision.indicators.macd,
        bb: decision.indicators.bb,
        stoch: decision.indicators.stoch,
        volume: decision.indicators.volume,
      },
    });
  } catch (error) {
    console.error("Error:", error);
    res.status(500).json({ error: "Analysis failed", message: error.message });
  }
};

export const simulateTrade = async (req, res) => {
  try {
    if (!req.body.symbol) {
      return res.status(400).json({ error: "Symbol is required in JSON body" });
    }
    const symbol = formatSymbol(req.body.symbol);
    const interval = req.body.interval || config.DEFAULT_INTERVAL;
    const initialCapital =
      parseInt(req.body.capital) || config.SAHAM.DEFAULT_CAPITAL;
    const maxLossInput = req.body.maxLoss || req.body.maxloss;
    const maxLossPercent = maxLossInput ? parseFloat(maxLossInput) : null;

    console.log(`\n🧪 Simulate triggered for ${symbol} [${interval}] ...`);

    const analysis = await analyzeSahamSymbol(
      symbol,
      interval,
      initialCapital,
      maxLossPercent
    );
    const safety = checkSafetyRules(analysis, "SAHAM", true);

    if (!safety.safe) {
      return res.status(403).json({
        error: "Simulation rejected by Executor Rules",
        reasons: safety.reasons,
      });
    }

    const signalPayload = {
      symbol,
      assetType: "SAHAM",
      interval, // INFORMATIONAL ONLY
      signal: analysis.signalResult.signal,
      entryPrice: analysis.quote.price,
      confidence: analysis.signalResult.confidence,
      score: analysis.signalResult.score,
      strength: analysis.signalResult.strength,
      tp: analysis.tpsl.tp,
      sl: analysis.tpsl.sl,
      allocatedAmount:
        analysis.moneyMgmt?.alokasiDana ||
        analysis.moneyMgmt?.priceLot?.totalBelanja,
    };

    if (signalPayload.signal === "SELL") {
      closePendingSignal(symbol, signalPayload.entryPrice);
    } else if (signalPayload.signal === "BUY") {
      logSignal(signalPayload);
    }

    res.json({
      success: true,
      message: "Saham simulation logged",
      safety,
      trade_plan: analysis.signalResult,
    });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
};

export const rawData = async (req, res) => {
  try {
    if (!req.query.symbol) {
      return res.status(400).json({ error: "Symbol parameter is required" });
    }
    const symbol = formatSymbol(req.query.symbol);
    const interval = req.query.interval || config.DEFAULT_INTERVAL;
    const timeframes = TIMEFRAME_MAP[interval] || TIMEFRAME_MAP["1d"];
    const count = parseInt(req.query.count) || 10;

    console.log(`\n📦 Fetching raw data for ${symbol} [${interval}]...`);

    const [multiTfData, ihsgData] = await Promise.all([
      fetchMultiTimeframe(symbol, timeframes),
      fetchIHSG(),
    ]);

    const formatCandle = (candle) => {
      return {
        date: candle.date, // WIB String
        o: candle.open,
        h: candle.high,
        l: candle.low,
        c: candle.close,
        v: candle.volume,
      };
    };

    const raw = {};
    timeframes.forEach((tf) => {
      if (multiTfData[tf]) {
        raw[tf] = multiTfData[tf].slice(-count).reverse().map(formatCandle);
      }
    });

    const ihsg = {
      "1h": [],
      "1d": ihsgData.slice(-count).reverse().map(formatCandle),
    };

    console.log(`✅ Raw data fetched for ${symbol}`);
    res.json({
      symbol,
      timestamp: getCurrentWIB(),
      raw,
      ihsg,
    });
  } catch (error) {
    console.error("Error fetching raw data:", error);
    res.status(500).json({ error: error.message });
  }
};

export const getSignalBasic = async (req, res) => {
  try {
    if (!req.query.symbol) {
      return res.status(400).json({ error: "Symbol parameter is required" });
    }
    const symbol = formatSymbol(req.query.symbol);
    const interval = req.query.interval || config.DEFAULT_INTERVAL;
    const timeframes = TIMEFRAME_MAP[interval] || TIMEFRAME_MAP["1d"];

    const [multiTfData, ihsgData, quote] = await Promise.all([
      fetchMultiTimeframe(symbol, timeframes),
      fetchIHSG(),
      getCurrentPrice(symbol),
    ]);
    const primaryData = multiTfData[interval];
    if (!primaryData || primaryData.length < 50)
      return res
        .status(400)
        .json({ error: "Insufficient data for " + interval });

    const ihsgAnalysis = analyzeIHSG(ihsgData);
    const decision = makeDecision(
      primaryData,
      multiTfData,
      ihsgAnalysis,
      timeframes
    );
    const signalResult = generateSignal(decision, quote.price, symbol);
    const tpsl = calculateTPSL(primaryData, signalResult.signal, quote.price);

    res.json({
      symbol,
      timestamp: getCurrentWIB(),
      currentPrice: quote.price,
      trade_plan: {
        valid: signalResult.signal === "BUY",
        signal: signalResult.signal,
        strength: signalResult.strength,
        confidence: signalResult.confidence,
        entry: signalResult.entryZone,
        tp: tpsl.tp,
        sl: tpsl.sl,
      },
      patterns: decision.patterns,
      ihsgTrend: ihsgAnalysis.trend,
      timeframeAlignment: decision.multiTimeframe.alignment,
      reasoning: signalResult.reasoning.slice(0, 5),
    });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
};

export const getSummaryStats = (req, res) => {
  const initialCapital =
    parseInt(req.query.capital) || config.SAHAM.DEFAULT_CAPITAL;
  const capitalStatus = getCapitalStatus("SAHAM", initialCapital);
  res.json({ ...getSignalSummary("SAHAM"), capitalStatus });
};

export const getHistoryLogs = (req, res) => {
  const { symbol, limit } = req.query;
  res.json(
    getSignalHistory({
      assetType: "SAHAM",
      symbol: symbol?.toUpperCase(),
      limit: parseInt(limit) || 50,
    })
  );
};
