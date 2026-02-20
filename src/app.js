import express from "express";
import cors from "cors";
import config from "./config.js";
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
} from "./shared/signalLogger.js";
import {
  openPosition,
  closePosition,
  getOpenPositions,
  getPositionHistory,
  getPositionSummary,
} from "./shared/positionTracker.js";

const app = express();
app.use(cors());
app.use(express.json());

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

app.get("/saham/analyze", async (req, res) => {
  try {
    if (!req.query.symbol) {
      return res.status(400).json({ error: "Symbol parameter is required" });
    }
    const symbol = formatSymbol(req.query.symbol);
    const initialCapital =
      parseInt(req.query.capital) || config.SAHAM.DEFAULT_CAPITAL;
    const capitalStatus = getCapitalStatus("SAHAM", initialCapital);
    const portfolio = {
      totalCapital: capitalStatus.available,
      maxLossPercent: parseFloat(req.query.maxLoss) || 1, // Default 1% max loss
      currentPositions: capitalStatus.openPositions,
    };
    console.log(`\n📊 Analyzing ${symbol}...`);

    const [multiTfData, ihsgData, quote] = await Promise.all([
      fetchMultiTimeframe(symbol),
      fetchIHSG(),
      getCurrentPrice(symbol),
    ]);
    const dailyData = multiTfData["1D"];
    if (!dailyData || dailyData.length < 50)
      return res.status(400).json({ error: "Insufficient data" });

    const ihsgAnalysis = analyzeIHSG(ihsgData);
    const correlation = calculateCorrelation(dailyData, ihsgData);
    const decision = makeDecision(dailyData, multiTfData, ihsgAnalysis);
    const signalResult = generateSignal(decision, quote.price, symbol);
    const tpsl = calculateTPSL(dailyData, signalResult.signal, quote.price);

    // Calculate trend strength from decision confidence
    const trendStrength = Math.abs(decision.score);

    const moneyMgmt = calculateMoneyManagement(
      portfolio,
      quote.price,
      tpsl.sl?.price,
      tpsl,
      signalResult.signal,
      trendStrength
    );

    console.log(`✅ ${signalResult.signal} (${signalResult.confidence}%)`);

    // Auto-log signal for performance tracking (BUY only — saham is long-only)
    if (signalResult.signal === "BUY") {
      logSignal({
        symbol,
        assetType: "SAHAM",
        signal: signalResult.signal,
        entryPrice: quote.price,
        confidence: signalResult.confidence,
        score: signalResult.score,
        strength: signalResult.strength,
        tp: tpsl.tp,
        sl: tpsl.sl,
        riskReward: tpsl.riskReward,
        timeframeAlignment: decision.multiTimeframe.alignment,
        marketTrend: ihsgAnalysis.trend,
        allocatedAmount: moneyMgmt.recommendation?.positionValue || 0,
      });
    } else if (signalResult.signal === "SELL") {
      // SELL = exit existing BUY (no new entry — saham can't short)
      closePendingSignal(symbol, quote.price);
    }

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
        // confidence: signalResult.confidence,
        // score: signalResult.score,
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
    const [multiTfData, ihsgData, quote] = await Promise.all([
      fetchMultiTimeframe(symbol),
      fetchIHSG(),
      getCurrentPrice(symbol),
    ]);
    const dailyData = multiTfData["1D"];
    if (!dailyData || dailyData.length < 50)
      return res.status(400).json({ error: "Insufficient data" });

    const ihsgAnalysis = analyzeIHSG(ihsgData);
    const decision = makeDecision(dailyData, multiTfData, ihsgAnalysis);
    const signalResult = generateSignal(decision, quote.price, symbol);
    const tpsl = calculateTPSL(dailyData, signalResult.signal, quote.price);

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
    const count = parseInt(req.query.count) || 10;

    console.log(`\n📦 Fetching raw data for ${symbol}...`);

    const [multiTfData, ihsgData] = await Promise.all([
      fetchMultiTimeframe(symbol),
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
    const raw = {
      "15m": multiTfData["15m"].slice(-count).reverse().map(formatCandle),
      "1h": multiTfData["1h"].slice(-count).reverse().map(formatCandle),
      "4h": multiTfData["4h"].slice(-count).reverse().map(formatCandle),
      "1d": multiTfData["1D"].slice(-count).reverse().map(formatCandle),
    };

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

app.get("/crypto/analyze", async (req, res) => {
  try {
    if (!req.query.symbol) {
      return res.status(400).json({ error: "Symbol parameter is required" });
    }
    const symbol = req.query.symbol.toUpperCase();
    const initialCapital =
      parseInt(req.query.capital) || config.CRYPTO.DEFAULT_CAPITAL;
    const capitalStatus = getCapitalStatus("CRYPTO", initialCapital);
    const portfolio = {
      totalCapital: capitalStatus.available,
      maxLossPercent: parseFloat(req.query.maxLoss) || 1,
      currentPositions: capitalStatus.openPositions,
    };
    const leverage = req.query.leverage ? parseInt(req.query.leverage) : null;

    console.log(`\n💎 Analyzing Crypto ${symbol}...`);

    // Fetch data in parallel
    const [multiTfData, btcDomData, quote] = await Promise.all([
      fetchBinanceMultiTf(symbol),
      fetchBTCDominance(),
      getBinancePrice(symbol),
    ]);

    // Use 1H data as primary (for H1 trading)
    const hourlyData = multiTfData["1h"];
    if (!hourlyData || hourlyData.length < 30)
      return res.status(400).json({ error: "Insufficient 1H data" });

    // BTC Market Sentiment (replaces IHSG for crypto)
    const btcMarket = analyzeBTCMarket(btcDomData, symbol);

    // Crypto Decision Engine (H1-optimized)
    const decision = makeCryptoDecision(hourlyData, multiTfData, btcMarket);

    // Generate Signal
    const signalResult = generateSignal(
      decision,
      quote.price,
      symbol,
      "BTC Market"
    );

    // TP/SL with crypto precision
    const tpsl = calculateTPSL(
      hourlyData,
      signalResult.signal,
      quote.price,
      "CRYPTO"
    );

    // Trend strength from decision confidence
    const trendStrength = Math.abs(decision.score);

    // Money Management
    const moneyMgmt = calculateMoneyManagement(
      portfolio,
      quote.price,
      tpsl.sl?.price,
      tpsl,
      signalResult.signal,
      trendStrength,
      "CRYPTO"
    );

    console.log(`✅ ${signalResult.signal} (${signalResult.confidence}%)`);

    // Auto-log signal for performance tracking
    if (signalResult.signal === "BUY" || signalResult.signal === "SELL") {
      logSignal({
        symbol,
        assetType: "CRYPTO",
        signal: signalResult.signal,
        entryPrice: quote.price,
        confidence: signalResult.confidence,
        score: signalResult.score,
        strength: signalResult.strength,
        tp: tpsl.tp,
        sl: tpsl.sl,
        riskReward: tpsl.riskReward,
        timeframeAlignment: decision.multiTimeframe.alignment,
        marketTrend: btcMarket.trend,
        allocatedAmount: moneyMgmt.recommendation?.positionValue || 0,
      });
    }

    // Update outcomes of pending crypto signals
    updateOutcomes(async (sym) => {
      try {
        const q = await getBinancePrice(sym);
        return q.price;
      } catch {
        return null;
      }
    }, "CRYPTO").catch(() => {});

    res.json({
      symbol,
      timestamp: getCurrentWIB(),
      capitalStatus,
      currentPrice: quote.price,
      change: quote.change,
      changePercent: quote.changePercent,
      volume: quote.volume,
      trade_plan: {
        valid: signalResult.signal === "BUY" || signalResult.signal === "SELL",
        signal: signalResult.signal,
        strength: signalResult.strength,
        // confidence: signalResult.confidence,
        // score: signalResult.score,
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
        index: "BTC Market",
        trend: btcMarket.trend,
        strength: btcMarket.strength,
        change1d: btcMarket.change1d,
        change7d: btcMarket.change7d,
        isCrash: btcMarket.isCrash,
        details: btcMarket.details,
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
      futures:
        signalResult.signal !== "WAIT"
          ? calculateFuturesPlan({
              capital: capitalStatus.available,
              entryPrice: quote.price,
              slPrice: tpsl.sl?.price,
              side: signalResult.signal === "BUY" ? "LONG" : "SHORT",
              leverage,
              tpsl,
            })
          : null,
    });
  } catch (error) {
    console.error("Crypto Analysis Error:", error);
    res.status(500).json({ error: "Analysis failed", message: error.message });
  }
});

app.get("/crypto/raw", async (req, res) => {
  try {
    if (!req.query.symbol)
      return res.status(400).json({ error: "Symbol required" });
    const symbol = req.query.symbol.toUpperCase();
    const count = parseInt(req.query.count) || 20;

    console.log(`\n📦 Fetching raw crypto data for ${symbol}...`);

    const [multiTfData, btcDomData] = await Promise.all([
      fetchBinanceMultiTf(symbol),
      fetchBTCDominance(),
    ]);

    // Helper (reuse logic or duplicate)
    // We can just send data as is because it's already formatted by binanceData.js
    const sliceData = (data) => data.slice(-count).reverse();

    res.json({
      symbol,
      timestamp: getCurrentWIB(),
      raw: {
        "15m": sliceData(multiTfData["15m"]),
        "1h": sliceData(multiTfData["1h"]),
        "4h": sliceData(multiTfData["4h"]),
        "1d": sliceData(multiTfData["1D"]),
      },
      btc_dominance: sliceData(btcDomData),
    });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// --- SIGNAL LOG ENDPOINTS ---

app.get("/crypto/signals/summary", (req, res) => {
  const initialCapital =
    parseInt(req.query.capital) || config.CRYPTO.DEFAULT_CAPITAL;
  const capitalStatus = getCapitalStatus("CRYPTO", initialCapital);
  res.json({ ...getSignalSummary("CRYPTO"), capitalStatus });
});

app.get("/crypto/signals/history", (req, res) => {
  const { symbol, limit } = req.query;
  res.json(
    getSignalHistory({
      assetType: "CRYPTO",
      symbol: symbol?.toUpperCase(),
      limit: parseInt(limit) || 50,
    })
  );
});

app.get("/saham/signals/summary", (req, res) => {
  const initialCapital =
    parseInt(req.query.capital) || config.SAHAM.DEFAULT_CAPITAL;
  const capitalStatus = getCapitalStatus("SAHAM", initialCapital);
  res.json({ ...getSignalSummary("SAHAM"), capitalStatus });
});

app.get("/saham/signals/history", (req, res) => {
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
║  GET  /saham/signals/summary         Signal perf stats       ║
║  POST /saham/position/open           Open position           ║
║  POST /saham/position/close          Close position          ║
║  GET  /saham/positions               Active positions        ║
║                                                              ║
║  CRYPTO:                                                     ║
║  GET  /crypto/analyze?symbol=PAIR    Full analysis           ║
║  GET  /crypto/signals/summary        Signal perf stats       ║
║  POST /crypto/position/open          Open position           ║
║  POST /crypto/position/close         Close position          ║
║  GET  /crypto/positions              Active positions        ║
╚══════════════════════════════════════════════════════════════╝
`)
);

export default app;
