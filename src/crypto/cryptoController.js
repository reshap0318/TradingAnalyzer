import config, { TIMEFRAME_MAP } from "../config.js";
import { analyzeCryptoSymbol } from "./cryptoAnalysis.js";
import {
  executeTrade,
  getTradingStatus,
  checkSafetyRules,
} from "../shared/tradeExecutor.js";
import {
  logSignal,
  updateOutcomes,
  getSummary as getSignalSummary,
  getHistory as getSignalHistory,
  getActive as getSignalActive,
  getCapitalStatus,
} from "../shared/signalLogger.js";
import {
  fetchMultiTimeframe as fetchBinanceMultiTf,
  fetchBTCDominance,
  getCurrentPrice as getBinancePrice,
} from "./binanceData.js";

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

export const analyzeCrypto = async (req, res) => {
  try {
    if (!req.query.symbol)
      return res.status(400).json({ error: "Symbol parameter is required" });

    const symbol = req.query.symbol.toUpperCase();
    const interval = req.query.interval || config.DEFAULT_INTERVAL;
    const initialCapital =
      parseInt(req.query.capital) || config.CRYPTO.DEFAULT_CAPITAL;
    const maxLossPercent = req.query.maxloss
      ? parseFloat(req.query.maxloss)
      : null;
    const leverage = req.query.leverage
      ? parseInt(req.query.leverage)
      : config.AUTO_TRADING.DEFAULT_LEVERAGE;

    console.log(`\n💎 Analyzing Crypto ${symbol} [${interval}]...`);
    const analysis = await analyzeCryptoSymbol(
      symbol,
      interval,
      leverage,
      initialCapital,
      maxLossPercent
    );

    console.log(
      `✅ ${analysis.signalResult.signal} (${analysis.signalResult.confidence}%)`
    );

    // Update outcomes of pending crypto signals passively (legacy logic)
    updateOutcomes(async (sym) => {
      try {
        const q = await getBinancePrice(sym);
        return q.price;
      } catch {
        return null;
      }
    }, "CRYPTO").catch(() => {});

    res.json({
      symbol: analysis.symbol,
      interval: analysis.interval,
      timestamp: getCurrentWIB(),
      capitalStatus: analysis.capitalStatus,
      currentPrice: analysis.quote.price,
      change: analysis.quote.change,
      changePercent: analysis.quote.changePercent,
      volume: analysis.quote.volume,
      trade_plan: {
        valid:
          analysis.signalResult.signal === "BUY" ||
          analysis.signalResult.signal === "SELL",
        signal: analysis.signalResult.signal,
        strength: analysis.signalResult.strength,
        entry: analysis.signalResult.entryZone,
        tp: analysis.tpsl.tp,
        sl: analysis.tpsl.sl,
        riskReward: analysis.tpsl.riskReward,
      },
      scoring: {
        totalScore: analysis.signalResult.score,
        confidence: analysis.signalResult.confidence,
        breakdown: analysis.decision.breakdown || [],
      },
      market_sentiment: {
        index: "BTC Market",
        trend: analysis.btcMarket.trend,
        strength: analysis.btcMarket.strength,
        change1d: analysis.btcMarket.change1d,
        change7d: analysis.btcMarket.change7d,
        isCrash: analysis.btcMarket.isCrash,
        details: analysis.btcMarket.details,
      },
      timeframes: Object.fromEntries(
        Object.entries(analysis.decision.multiTimeframe.timeframes).map(
          ([tf, d]) => [tf, { trend: d.trend, signal: Math.round(d.signal) }]
        )
      ),
      timeframeAlignment: analysis.decision.multiTimeframe.alignment,
      reasoning: analysis.signalResult.reasoning,
      warnings: analysis.signalResult.warnings,
      moneyManagement: analysis.moneyMgmt,
      patterns: analysis.decision.patterns,
      indicators: {
        ma: analysis.decision.indicators.ma,
        rsi: analysis.decision.indicators.rsi,
        macd: analysis.decision.indicators.macd,
        bb: analysis.decision.indicators.bb,
        stoch: analysis.decision.indicators.stoch,
        volume: analysis.decision.indicators.volume,
      },
      binance_execution: analysis.futures?.binance_execution || null,
    });
  } catch (error) {
    console.error("Error:", error);
    res.status(500).json({ error: "Analysis failed", message: error.message });
  }
};

export const tradeStatus = (req, res) => {
  res.json(getTradingStatus("CRYPTO"));
};

export const tradeStop = (req, res) => {
  config.AUTO_TRADING.ENABLED = false;
  res.json({
    stopped: true,
    message: req.body.reason || "Manual killswitch activated.",
  });
};

export const tradeStart = (req, res) => {
  config.AUTO_TRADING.ENABLED = true;
  res.json({
    started: true,
    message: "Auto-trading enabled.",
  });
};

export const tradeAuto = async (req, res) => {
  try {
    if (!config.AUTO_TRADING.ENABLED) {
      return res
        .status(403)
        .json({ error: "Auto-Trading is disabled globally in config.js." });
    }
    if (!req.body.symbol) {
      return res.status(400).json({ error: "Symbol is required in JSON body" });
    }

    const symbol = req.body.symbol.toUpperCase();
    const interval = req.body.interval || config.DEFAULT_INTERVAL;
    const initialCapital =
      parseInt(req.body.capital) || config.CRYPTO.DEFAULT_CAPITAL;
    const maxLossPercent = req.body.maxloss
      ? parseFloat(req.body.maxloss)
      : null;
    const leverage = req.body.leverage || config.AUTO_TRADING.DEFAULT_LEVERAGE;

    console.log(`\n🤖 AutoTrader triggered for ${symbol} [${interval}] ...`);

    const analysis = await analyzeCryptoSymbol(
      symbol,
      interval,
      leverage,
      initialCapital,
      maxLossPercent
    );
    const result = await executeTrade(analysis);

    res.json({
      symbol,
      interval,
      executed: result.executed,
      network: result.network,
      reasons: result.reasons, // If blocked
      trade_plan: analysis.signalResult,
      capitalStatus: analysis.capitalStatus,
      execution_details: {
        order: result.order,
        tpSlTrailing: result.tpSlOrders,
      },
    });
  } catch (error) {
    console.error("Auto Trading API Error:", error);
    res.status(500).json({ error: "Execution failed", message: error.message });
  }
};

export const simulateTrade = async (req, res) => {
  try {
    if (!req.body.symbol) {
      return res.status(400).json({ error: "Symbol is required in JSON body" });
    }
    const symbol = req.body.symbol.toUpperCase();
    const interval = req.body.interval || config.DEFAULT_INTERVAL;
    const initialCapital =
      parseInt(req.body.capital) || config.CRYPTO.DEFAULT_CAPITAL;
    const maxLossPercent = req.body.maxloss
      ? parseFloat(req.body.maxloss)
      : null;
    const leverage =
      parseInt(req.body.leverage) || config.AUTO_TRADING.DEFAULT_LEVERAGE;

    console.log(`\n🧪 Simulate triggered for ${symbol} [${interval}] ...`);

    const analysis = await analyzeCryptoSymbol(
      symbol,
      interval,
      leverage,
      initialCapital,
      maxLossPercent
    );
    const safety = checkSafetyRules(analysis, "CRYPTO", true);

    if (!safety.safe) {
      return res.status(403).json({
        error: "Simulation rejected by Executor Rules",
        reasons: safety.reasons,
      });
    }

    const signalPayload = {
      symbol,
      assetType: "CRYPTO",
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
      binance_execution: analysis.futures?.binance_execution || null,
    };

    if (signalPayload.signal !== "WAIT") {
      logSignal(signalPayload);
    }

    res.json({
      success: true,
      message: "Crypto simulation logged",
      safety,
      trade_plan: analysis.signalResult,
      binance_execution: signalPayload.binance_execution,
    });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
};

export const rawData = async (req, res) => {
  try {
    if (!req.query.symbol)
      return res.status(400).json({ error: "Symbol required" });
    const symbol = req.query.symbol.toUpperCase();
    const interval = req.query.interval || config.DEFAULT_INTERVAL;
    const timeframes = TIMEFRAME_MAP[interval] || TIMEFRAME_MAP["1d"];
    const count = parseInt(req.query.count) || 20;

    console.log(`\n📦 Fetching raw crypto data for ${symbol} [${interval}]...`);

    const [multiTfData, btcDomData] = await Promise.all([
      fetchBinanceMultiTf(symbol, timeframes),
      fetchBTCDominance(),
    ]);

    const sliceData = (data) => data.slice(-count).reverse();

    const raw = {};
    timeframes.forEach((tf) => {
      if (multiTfData[tf]) {
        raw[tf] = sliceData(multiTfData[tf]);
      }
    });

    res.json({
      symbol,
      timestamp: getCurrentWIB(),
      raw,
      btc_dominance: sliceData(btcDomData),
    });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
};

export const getSummaryStats = (req, res) => {
  const initialCapital =
    parseInt(req.query.capital) || config.CRYPTO.DEFAULT_CAPITAL;
  const capitalStatus = getCapitalStatus("CRYPTO", initialCapital);
  res.json({ ...getSignalSummary("CRYPTO"), capitalStatus });
};

export const getHistoryLogs = (req, res) => {
  const { symbol, limit } = req.query;
  res.json(
    getSignalHistory({
      assetType: "CRYPTO",
      symbol: symbol?.toUpperCase(),
      limit: parseInt(limit) || 50,
    })
  );
};

export const getActiveLogs = (req, res) => {
  const { symbol, limit } = req.query;
  res.json(
    getSignalActive({
      assetType: "CRYPTO",
      symbol: symbol?.toUpperCase(),
      limit: parseInt(limit) || 50,
    })
  );
};
