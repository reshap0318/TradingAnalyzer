import express from "express";
import cors from "cors";
import config, { TIMEFRAME_MAP } from "./config.js";
import {
  fetchMultiTimeframe,
  fetchIHSG,
  getCurrentPrice,
} from "./saham/yahooFinance.js";
import {
  fetchMultiTimeframe as fetchBinanceMultiTf,
  fetchBTCDominance,
  getCurrentPrice as getBinancePrice,
} from "./crypto/binanceData.js";
import { analyzeIHSG, calculateCorrelation } from "./saham/ihsgAnalyzer.js";
import { analyzeBTCMarket } from "./crypto/btcMarketAnalyzer.js";
import { makeDecision } from "./saham/decisionEngine.js";
import { makeCryptoDecision } from "./crypto/cryptoDecisionEngine.js";
import { calculateFuturesPlan } from "./crypto/futuresCalculator.js";
import { generateSignal } from "./shared/signalGenerator.js";
import { calculateTPSL } from "./shared/tpslCalculator.js";
import { calculateMoneyManagement } from "./shared/moneyManagement.js";
import {
  logSignal,
  closePendingSignal,
  updateOutcomes,
  getSummary as getSignalSummary,
  getHistory as getSignalHistory,
  getCapitalStatus,
  getSummary,
  getHistory,
} from "./shared/signalLogger.js";
import { executeTrade, getTradingStatus } from "./shared/tradeExecutor.js";
import { initClient as initBinanceTrader } from "./crypto/binanceTrader.js";

const app = express();
app.use(cors());
app.use(express.json());

// Initialize optional executor services
if (config.AUTO_TRADING.ENABLED) {
  initBinanceTrader(config.AUTO_TRADING.USE_TESTNET).catch((e) =>
    console.log("Init node skipped:", e.message)
  );
}

const dataDir = path.join(__dirname, "../data");

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

export async function analyzeSahamSymbol(
  symbol,
  interval,
  initialCapital = config.SAHAM.DEFAULT_CAPITAL
) {
  const timeframes = TIMEFRAME_MAP[interval] || TIMEFRAME_MAP["1d"];
  const capitalStatus = getCapitalStatus("SAHAM", initialCapital);
  const portfolio = {
    totalCapital: capitalStatus.available,
    maxLossPercent: config.SAHAM.MONEY_MANAGEMENT.MAX_RISK_PER_TRADE * 100,
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

app.get("/saham/analyze", async (req, res) => {
  try {
    if (!req.query.symbol) {
      return res.status(400).json({ error: "Symbol parameter is required" });
    }
    const symbol = formatSymbol(req.query.symbol);
    const interval = req.query.interval || config.DEFAULT_INTERVAL;
    const initialCapital =
      parseInt(req.query.capital) || config.SAHAM.DEFAULT_CAPITAL;

    console.log(`\n📊 Analyzing ${symbol} [${interval}]...`);

    const analysis = await analyzeSahamSymbol(symbol, interval, initialCapital);
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
});

app.get("/saham/signal", async (req, res) => {
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
      patterns: decision.patterns, // Added patterns output
      ihsgTrend: ihsgAnalysis.trend,
      timeframeAlignment: decision.multiTimeframe.alignment,
      reasoning: signalResult.reasoning.slice(0, 5),
    });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// Raw data endpoint - returns last 10 candles for each timeframe
app.get("/saham/raw", async (req, res) => {
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

    // Format OHLCV data to be more compact
    const formatCandle = (candle) => {
      // Data Layer now returns WIB date string in 'date' field
      return {
        date: candle.date, // WIB String
        o: candle.open,
        h: candle.high,
        l: candle.low,
        c: candle.close,
        v: candle.volume,
      };
    };

    // Get last N candles for each timeframe
    const raw = {};
    timeframes.forEach((tf) => {
      if (multiTfData[tf]) {
        raw[tf] = multiTfData[tf].slice(-count).reverse().map(formatCandle);
      }
    });

    // Fetch IHSG hourly data (none)
    const getDateStr = (daysAgo) => {
      const d = new Date();
      d.setDate(d.getDate() - daysAgo);
      return d.toISOString().split("T")[0];
    };

    // IHSG data
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
});

// --- CRYPTO ENDPOINTS ---

export async function analyzeCryptoSymbol(
  symbol,
  interval,
  leverage,
  initialCapital = config.CRYPTO.DEFAULT_CAPITAL
) {
  const timeframes = TIMEFRAME_MAP[interval] || TIMEFRAME_MAP["1d"];
  const capitalStatus = getCapitalStatus("CRYPTO", initialCapital);
  const portfolio = {
    totalCapital: capitalStatus.available,
    maxLossPercent: config.CRYPTO.MONEY_MANAGEMENT.MAX_RISK_PER_TRADE * 100,
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

app.get("/crypto/analyze", async (req, res) => {
  try {
    if (!req.query.symbol)
      return res.status(400).json({ error: "Symbol parameter is required" });

    const symbol = req.query.symbol.toUpperCase();
    const interval = req.query.interval || config.DEFAULT_INTERVAL;
    const initialCapital =
      parseInt(req.query.capital) || config.CRYPTO.DEFAULT_CAPITAL;
    const leverage = req.query.leverage ? parseInt(req.query.leverage) : null;

    console.log(`\n💎 Analyzing Crypto ${symbol} [${interval}]...`);
    const analysis = await analyzeCryptoSymbol(
      symbol,
      interval,
      leverage,
      initialCapital
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
      futures:
        analysis.signalResult.signal !== "WAIT"
          ? calculateFuturesPlan({
              capital: analysis.capitalStatus.available,
              entryPrice: analysis.quote.price,
              slPrice: analysis.tpsl.sl?.price,
              side: analysis.signalResult.signal === "BUY" ? "LONG" : "SHORT",
              leverage: analysis.leverage,
              tpsl: analysis.tpsl,
            })
          : null,
    });
  } catch (error) {
    console.error("Crypto Analysis Error:", error);
    res.status(500).json({ error: "Analysis failed", message: error.message });
  }
});

// --- CRYPTO AUTO TRADING EXECUTOR ---

app.get("/crypto/trade/status", (req, res) => {
  res.json(getTradingStatus());
});

app.post("/crypto/trade/stop", (req, res) => {
  config.AUTO_TRADING.ENABLED = false;
  res.json({
    stopped: true,
    message: req.body.reason || "Manual killswitch activated.",
  });
});

app.post("/crypto/trade/auto", async (req, res) => {
  try {
    if (!config.AUTO_TRADING.ENABLED) {
      return res
        .status(403)
        .json({ error: "Auto-Trading is disabled globally in config.js." });
    }
    if (!req.body.symbol) {
      return res.status(400).json({
        error: "Symbol is required in JSON body (e.g. { symbol: 'BTCUSDT' })",
      });
    }

    const symbol = req.body.symbol.toUpperCase();
    const interval = req.body.interval || config.DEFAULT_INTERVAL;
    const initialCapital =
      parseInt(req.body.capital) || config.CRYPTO.DEFAULT_CAPITAL;
    const mode = req.body.mode || config.AUTO_TRADING.MODE;
    const leverage = req.body.leverage || config.AUTO_TRADING.DEFAULT_LEVERAGE;

    if (mode === "paper") {
      return res.status(400).json({
        error:
          "Paper mode has been migrated. Please use /crypto/trade/simulate instead.",
      });
    }

    console.log(`\n🤖 AutoTrader triggered for ${symbol} [${interval}] ...`);

    // 1. Analyze Core Logic (no API response formatted)
    const analysis = await analyzeCryptoSymbol(
      symbol,
      interval,
      leverage,
      initialCapital
    );

    // 2. Pass to Executor Layer
    const result = await executeTrade(analysis, mode);

    res.json({
      symbol,
      interval,
      executed: result.executed,
      mode: result.mode,
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
});

app.get("/crypto/raw", async (req, res) => {
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

    // Helper (reuse logic or duplicate)
    // We can just send data as is because it's already formatted by binanceData.js
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
});

// --- SIMULATION ENDPOINTS (Replaces Old Passive Logs) ---
app.post("/saham/trade/simulate", async (req, res) => {
  try {
    if (!req.body.symbol) {
      return res.status(400).json({ error: "Symbol is required in JSON body" });
    }
    const symbol = formatSymbol(req.body.symbol);
    const interval = req.body.interval || config.DEFAULT_INTERVAL;
    const initialCapital =
      parseInt(req.body.capital) || config.SAHAM.DEFAULT_CAPITAL;

    console.log(`\n🧪 Simulate triggered for ${symbol} [${interval}] ...`);

    const analysis = await analyzeSahamSymbol(symbol, interval, initialCapital);
    const safety = checkSafetyRules(analysis, "SAHAM");

    if (!safety.safe) {
      return res.status(403).json({
        error: "Simulation rejected by Executor Rules",
        reasons: safety.reasons,
      });
    }

    const signalPayload = {
      symbol,
      assetType: "SAHAM",
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
});

app.post("/crypto/trade/simulate", async (req, res) => {
  try {
    if (!req.body.symbol) {
      return res.status(400).json({ error: "Symbol is required in JSON body" });
    }
    const symbol = req.body.symbol.toUpperCase();
    const interval = req.body.interval || config.DEFAULT_INTERVAL;
    const initialCapital =
      parseInt(req.body.capital) || config.CRYPTO.DEFAULT_CAPITAL;
    const leverage =
      parseInt(req.body.leverage) || config.AUTO_TRADING.DEFAULT_LEVERAGE;

    console.log(`\n🧪 Simulate triggered for ${symbol} [${interval}] ...`);

    const analysis = await analyzeCryptoSymbol(
      symbol,
      interval,
      leverage,
      initialCapital
    );
    const safety = checkSafetyRules(analysis, "CRYPTO");

    if (!safety.safe) {
      return res.status(403).json({
        error: "Simulation rejected by Executor Rules",
        reasons: safety.reasons,
      });
    }

    const signalPayload = {
      symbol,
      assetType: "CRYPTO",
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

    if (signalPayload.signal !== "WAIT") {
      logSignal(signalPayload);
    }

    res.json({
      success: true,
      message: "Crypto simulation logged",
      safety,
      trade_plan: analysis.signalResult,
    });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.get("/crypto/simulate/summary", (req, res) => {
  const initialCapital =
    parseInt(req.query.capital) || config.CRYPTO.DEFAULT_CAPITAL;
  const capitalStatus = getCapitalStatus("CRYPTO", initialCapital);
  res.json({ ...getSignalSummary("CRYPTO"), capitalStatus });
});

app.get("/crypto/simulate/history", (req, res) => {
  const { symbol, limit } = req.query;
  res.json(
    getSignalHistory({
      assetType: "CRYPTO",
      symbol: symbol?.toUpperCase(),
      limit: parseInt(limit) || 50,
    })
  );
});

app.get("/saham/simulate/summary", (req, res) => {
  const initialCapital =
    parseInt(req.query.capital) || config.SAHAM.DEFAULT_CAPITAL;
  const capitalStatus = getCapitalStatus("SAHAM", initialCapital);
  res.json({ ...getSignalSummary("SAHAM"), capitalStatus });
});

app.get("/saham/simulate/history", (req, res) => {
  const { symbol, limit } = req.query;
  res.json(
    getSignalHistory({
      assetType: "SAHAM",
      symbol: symbol?.toUpperCase(),
      limit: parseInt(limit) || 50,
    })
  );
});

// --- POSITION TRACKER ENDPOINTS ---

app.post("/crypto/position/open", (req, res) => {
  const { symbol, side, entryPrice, quantity, sl, tp, notes } = req.body;
  if (!symbol || !entryPrice || !quantity) {
    return res
      .status(400)
      .json({ error: "symbol, entryPrice, quantity are required" });
  }
  res.json(
    openPosition({
      symbol: symbol.toUpperCase(),
      assetType: "CRYPTO",
      side: side?.toUpperCase() || "LONG",
      entryPrice: parseFloat(entryPrice),
      quantity: parseFloat(quantity),
      sl: sl ? parseFloat(sl) : null,
      tp: tp ? parseFloat(tp) : null,
      notes: notes || "",
    })
  );
});

app.post("/crypto/position/close", (req, res) => {
  const { symbol, exitPrice, reason } = req.body;
  if (!symbol || !exitPrice) {
    return res.status(400).json({ error: "symbol, exitPrice are required" });
  }
  res.json(
    closePosition({
      symbol: symbol.toUpperCase(),
      assetType: "CRYPTO",
      exitPrice: parseFloat(exitPrice),
      reason: reason || "MANUAL",
    })
  );
});

app.get("/crypto/positions", (req, res) => {
  res.json(getOpenPositions("CRYPTO"));
});

app.get("/crypto/positions/history", (req, res) => {
  const { symbol, limit } = req.query;
  res.json(
    getPositionHistory({
      assetType: "CRYPTO",
      symbol: symbol?.toUpperCase(),
      limit: parseInt(limit) || 50,
    })
  );
});

app.get("/crypto/positions/summary", (req, res) => {
  res.json(getPositionSummary("CRYPTO"));
});

// --- SAHAM POSITION TRACKER ENDPOINTS ---

app.post("/saham/position/open", (req, res) => {
  const { symbol, side, entryPrice, quantity, sl, tp, notes } = req.body;
  if (!symbol || !entryPrice || !quantity) {
    return res
      .status(400)
      .json({ error: "symbol, entryPrice, quantity are required" });
  }
  const formatted = symbol.includes(".")
    ? symbol.toUpperCase()
    : `${symbol.toUpperCase()}.JK`;
  res.json(
    openPosition({
      symbol: formatted,
      assetType: "SAHAM",
      side: side?.toUpperCase() || "LONG",
      entryPrice: parseFloat(entryPrice),
      quantity: parseFloat(quantity),
      sl: sl ? parseFloat(sl) : null,
      tp: tp ? parseFloat(tp) : null,
      notes: notes || "",
    })
  );
});

app.post("/saham/position/close", (req, res) => {
  const { symbol, exitPrice, reason } = req.body;
  if (!symbol || !exitPrice) {
    return res.status(400).json({ error: "symbol, exitPrice are required" });
  }
  const formatted = symbol.includes(".")
    ? symbol.toUpperCase()
    : `${symbol.toUpperCase()}.JK`;
  res.json(
    closePosition({
      symbol: formatted,
      assetType: "SAHAM",
      exitPrice: parseFloat(exitPrice),
      reason: reason || "MANUAL",
    })
  );
});

app.get("/saham/positions", (req, res) => {
  res.json(getOpenPositions("SAHAM"));
});

app.get("/saham/positions/history", (req, res) => {
  const { symbol, limit } = req.query;
  const formatted = symbol
    ? symbol.includes(".")
      ? symbol.toUpperCase()
      : `${symbol.toUpperCase()}.JK`
    : undefined;
  res.json(
    getPositionHistory({
      assetType: "SAHAM",
      symbol: formatted,
      limit: parseInt(limit) || 50,
    })
  );
});

app.get("/saham/positions/summary", (req, res) => {
  res.json(getPositionSummary("SAHAM"));
});

// --- HEALTH ---

app.get("/health", (req, res) =>
  res.json({ status: "ok", timestamp: getCurrentWIB() })
);

app.listen(config.SERVER.PORT, "0.0.0.0", () =>
  console.log(`
╔══════════════════════════════════════════════════════════════╗
║            TRADING ANALYZER API - RUNNING                    ║
╠══════════════════════════════════════════════════════════════╣
║  Server: http://0.0.0.0:${config.SERVER.PORT}                              ║
║                                                              ║
║  SAHAM:                                                      ║
║  GET  /saham/analyze?symbol=CODE     Full analysis           ║
║  GET  /saham/signal?symbol=CODE      Quick signal            ║
║  POST /saham/signals/log             Log a signal manually   ║
║  GET  /saham/signals/summary         Signal perf stats       ║
║  POST /saham/position/open           Open position           ║
║  POST /saham/position/close          Close position          ║
║  GET  /saham/positions               Active positions        ║
║                                                              ║
║  CRYPTO:                                                     ║
║  GET  /crypto/analyze?symbol=PAIR    Full analysis           ║
║  POST /crypto/signals/log            Log a signal manually   ║
║  GET  /crypto/signals/summary        Signal perf stats       ║
║  POST /crypto/position/open          Open position           ║
║  POST /crypto/position/close         Close position          ║
║  GET  /crypto/positions              Active positions        ║
╚══════════════════════════════════════════════════════════════╝
`)
);

export default app;
