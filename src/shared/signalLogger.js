import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const LOG_FILE = path.join(__dirname, "../../data/signal_log.json");

/**
 * Signal Logger — Automatically records every BUY/SELL signal for performance evaluation.
 * Supports both CRYPTO and SAHAM asset types.
 *
 * Signals start as PENDING, then get updated when TP/SL is hit or they expire.
 * Expiry: 48h for crypto, 10 days for saham.
 */

function loadLog() {
  try {
    if (fs.existsSync(LOG_FILE)) {
      return JSON.parse(fs.readFileSync(LOG_FILE, "utf8"));
    }
  } catch (e) {
    console.error("Error loading signal log:", e.message);
  }
  return [];
}

function saveLog(data) {
  const dir = path.dirname(LOG_FILE);
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(LOG_FILE, JSON.stringify(data, null, 2));
}

/**
 * Log a new signal. Called automatically when analyze endpoint produces BUY/SELL.
 * Rules:
 *   - 1 PENDING signal per symbol max
 *   - Same direction as existing PENDING → skip (hold)
 *   - Opposite direction → close old as SIGNAL_REVERSED, create new entry
 */
export function logSignal({
  symbol,
  assetType = "CRYPTO",
  signal,
  entryPrice,
  confidence,
  score,
  strength,
  tp1,
  tp2,
  tp3,
  sl,
  riskReward,
  timeframeAlignment,
  marketTrend,
  allocatedAmount = 0,
}) {
  const log = loadLog();
  const now = new Date();

  // Find existing PENDING signal for this symbol
  const pendingIdx = log.findIndex(
    (entry) => entry.symbol === symbol && entry.outcome === "PENDING"
  );

  if (pendingIdx !== -1) {
    const pending = log[pendingIdx];

    // Same direction → skip (already tracking this signal)
    if (pending.signal === signal) {
      return {
        logged: false,
        reason: `Already tracking ${signal} for ${symbol} since ${pending.timestamp}`,
      };
    }

    // Opposite direction → close old as SIGNAL_REVERSED
    const isBuy = pending.signal === "BUY";
    const pnl = isBuy
      ? ((entryPrice - pending.entryPrice) / pending.entryPrice) * 100
      : ((pending.entryPrice - entryPrice) / pending.entryPrice) * 100;
    const ageHours = (now - new Date(pending.timestamp)) / (1000 * 60 * 60);

    pending.outcome = "SIGNAL_REVERSED";
    pending.exitPrice = entryPrice;
    pending.exitTime = now.toISOString();
    pending.pnlPercent = Math.round(pnl * 100) / 100;
    pending.pnlDollar = Math.round((pending.allocatedAmount || 0) * pnl) / 100;
    pending.holdHours = Math.round(ageHours * 10) / 10;

    const emoji = pnl >= 0 ? "✅" : "❌";
    console.log(
      `🔄 Signal reversed: ${pending.signal}→${signal} ${symbol} | PnL: ${
        pnl >= 0 ? "+" : ""
      }${pending.pnlPercent}%`
    );
  }

  const entry = {
    id: `${symbol}_${now.getTime()}`,
    symbol,
    assetType,
    signal,
    timestamp: now.toISOString(),
    entryPrice,
    confidence,
    score,
    strength,
    tp1: tp1?.price || null,
    tp2: tp2?.price || null,
    tp3: tp3?.price || null,
    sl: sl?.price || null,
    riskReward: riskReward?.tp1 || null,
    timeframeAlignment,
    marketTrend,
    allocatedAmount: Math.round(allocatedAmount * 100) / 100,
    // Outcome tracking
    outcome: "PENDING",
    highestPrice: entryPrice,
    lowestPrice: entryPrice,
    exitPrice: null,
    exitTime: null,
    pnlPercent: null,
    pnlDollar: null,
    holdHours: null,
  };

  log.push(entry);
  saveLog(log);
  console.log(`📝 Signal logged: ${symbol} ${signal} @ ${entryPrice}`);
  return { logged: true, id: entry.id };
}

/**
 * Close a PENDING signal as SIGNAL_REVERSED without creating a new entry.
 * Used for saham: SELL means "exit BUY" but doesn't open a short position.
 */
export function closePendingSignal(symbol, currentPrice) {
  const log = loadLog();
  const now = new Date();

  const pending = log.find(
    (entry) => entry.symbol === symbol && entry.outcome === "PENDING"
  );
  if (!pending) return { closed: false, reason: "No PENDING signal found" };

  const isBuy = pending.signal === "BUY";
  const pnl = isBuy
    ? ((currentPrice - pending.entryPrice) / pending.entryPrice) * 100
    : ((pending.entryPrice - currentPrice) / pending.entryPrice) * 100;
  const ageHours = (now - new Date(pending.timestamp)) / (1000 * 60 * 60);

  pending.outcome = "SIGNAL_REVERSED";
  pending.exitPrice = currentPrice;
  pending.exitTime = now.toISOString();
  pending.pnlPercent = Math.round(pnl * 100) / 100;
  pending.pnlDollar = Math.round((pending.allocatedAmount || 0) * pnl) / 100;
  pending.holdHours = Math.round(ageHours * 10) / 10;

  saveLog(log);
  console.log(
    `🔄 Signal closed (reversed): ${
      pending.signal
    } ${symbol} @ ${currentPrice} | PnL: ${pnl >= 0 ? "+" : ""}${
      pending.pnlPercent
    }%`
  );
  return { closed: true, pnlPercent: pending.pnlPercent };
}

/**
 * Update outcomes of PENDING signals by checking current prices.
 * Call this periodically (e.g., every hour via cron or on each analyze call).
 *
 * @param {Function} getPriceFn - async function(symbol) => number (current price)
 * @param {string} assetType - "CRYPTO" or "SAHAM" — only check signals of this type
 */
export async function updateOutcomes(getPriceFn, assetType = null) {
  const log = loadLog();
  let updated = 0;

  for (const entry of log) {
    if (entry.outcome !== "PENDING") continue;
    if (assetType && entry.assetType !== assetType) continue;

    const ageHours =
      (Date.now() - new Date(entry.timestamp).getTime()) / (1000 * 60 * 60);
    const maxAge = entry.assetType === "CRYPTO" ? 48 : 240; // 48h crypto, 10d saham

    // Expire old signals
    if (ageHours > maxAge) {
      entry.outcome = "EXPIRED";
      entry.holdHours = Math.round(ageHours);
      entry.pnlPercent = 0;
      entry.pnlDollar = 0;
      updated++;
      continue;
    }

    // Get current price
    let currentPrice;
    try {
      currentPrice = await getPriceFn(entry.symbol, entry.assetType);
      if (!currentPrice) continue;
    } catch {
      continue;
    }

    // Track high/low since entry
    entry.highestPrice = Math.max(
      entry.highestPrice || entry.entryPrice,
      currentPrice
    );
    entry.lowestPrice = Math.min(
      entry.lowestPrice || entry.entryPrice,
      currentPrice
    );

    // Check TP/SL hit
    const isBuy = entry.signal === "BUY";

    // Check SL first (priority)
    if (entry.sl) {
      const slHit = isBuy ? currentPrice <= entry.sl : currentPrice >= entry.sl;
      if (slHit) {
        entry.outcome = "SL_HIT";
        entry.exitPrice = entry.sl;
        entry.exitTime = new Date().toISOString();
        entry.pnlPercent = isBuy
          ? ((entry.sl - entry.entryPrice) / entry.entryPrice) * 100
          : ((entry.entryPrice - entry.sl) / entry.entryPrice) * 100;
        entry.pnlDollar =
          Math.round((entry.allocatedAmount || 0) * entry.pnlPercent) / 100;
        entry.holdHours = Math.round(ageHours);
        updated++;
        continue;
      }
    }

    // Check TP3 first (best outcome), then TP2, TP1
    const tpChecks = [
      { price: entry.tp3, label: "TP3_HIT" },
      { price: entry.tp2, label: "TP2_HIT" },
      { price: entry.tp1, label: "TP1_HIT" },
    ];

    for (const tp of tpChecks) {
      if (!tp.price) continue;
      const tpHit = isBuy ? currentPrice >= tp.price : currentPrice <= tp.price;
      if (tpHit) {
        entry.outcome = tp.label;
        entry.exitPrice = tp.price;
        entry.exitTime = new Date().toISOString();
        entry.pnlPercent = isBuy
          ? ((tp.price - entry.entryPrice) / entry.entryPrice) * 100
          : ((entry.entryPrice - tp.price) / entry.entryPrice) * 100;
        entry.pnlDollar =
          Math.round((entry.allocatedAmount || 0) * entry.pnlPercent) / 100;
        entry.holdHours = Math.round(ageHours);
        updated++;
        break;
      }
    }
  }

  if (updated > 0) {
    saveLog(log);
    console.log(`📊 Signal outcomes updated: ${updated} entries`);
  }
  return updated;
}

/**
 * Get available capital status. Calculates:
 * available = initialCapital - allocated(PENDING) + realized PnL
 *
 * @param {string} assetType - "CRYPTO" or "SAHAM"
 * @param {number} initialCapital - Starting capital from config or query param
 */
export function getCapitalStatus(assetType, initialCapital) {
  const log = loadLog();
  const filtered = log.filter((e) => e.assetType === assetType);

  const pending = filtered.filter((e) => e.outcome === "PENDING");
  const completed = filtered.filter((e) => e.outcome !== "PENDING");

  const allocated = pending.reduce(
    (sum, e) => sum + (e.allocatedAmount || 0),
    0
  );
  const realizedPnl = completed.reduce((sum, e) => sum + (e.pnlDollar || 0), 0);
  const available = initialCapital - allocated + realizedPnl;

  return {
    initialCapital: Math.round(initialCapital * 100) / 100,
    allocated: Math.round(allocated * 100) / 100,
    realizedPnl: Math.round(realizedPnl * 100) / 100,
    available: Math.round(available * 100) / 100,
    openPositions: pending.length,
  };
}

/**
 * Get performance summary, optionally filtered by asset type.
 */
export function getSummary(assetType = null) {
  const log = loadLog();
  const filtered = assetType
    ? log.filter((e) => e.assetType === assetType)
    : log;

  const completed = filtered.filter((e) => e.outcome !== "PENDING");
  const wins = completed.filter((e) =>
    ["TP1_HIT", "TP2_HIT", "TP3_HIT"].includes(e.outcome)
  );
  const losses = completed.filter((e) => e.outcome === "SL_HIT");
  const reversed = completed.filter((e) => e.outcome === "SIGNAL_REVERSED");
  const reversedWins = reversed.filter((e) => (e.pnlPercent || 0) > 0);
  const reversedLosses = reversed.filter((e) => (e.pnlPercent || 0) <= 0);
  const expired = completed.filter((e) => e.outcome === "EXPIRED");
  const pending = filtered.filter((e) => e.outcome === "PENDING");

  const allWins = [...wins, ...reversedWins];
  const allLosses = [...losses, ...reversedLosses];

  const totalPnl = completed.reduce((sum, e) => sum + (e.pnlPercent || 0), 0);
  const avgWin =
    allWins.length > 0
      ? allWins.reduce((sum, e) => sum + e.pnlPercent, 0) / allWins.length
      : 0;
  const avgLoss =
    allLosses.length > 0
      ? allLosses.reduce((sum, e) => sum + e.pnlPercent, 0) / allLosses.length
      : 0;
  const grossProfit = allWins.reduce((sum, e) => sum + (e.pnlPercent || 0), 0);
  const grossLoss = Math.abs(
    allLosses.reduce((sum, e) => sum + (e.pnlPercent || 0), 0)
  );

  // Dollar amounts
  const totalPnlDollar = completed.reduce(
    (sum, e) => sum + (e.pnlDollar || 0),
    0
  );
  const avgWinDollar =
    allWins.length > 0
      ? allWins.reduce((sum, e) => sum + (e.pnlDollar || 0), 0) / allWins.length
      : 0;
  const avgLossDollar =
    allLosses.length > 0
      ? allLosses.reduce((sum, e) => sum + (e.pnlDollar || 0), 0) /
        allLosses.length
      : 0;

  return {
    totalSignals: filtered.length,
    pending: pending.length,
    completed: completed.length,
    wins: allWins.length,
    losses: allLosses.length,
    reversed: reversed.length,
    expired: expired.length,
    winRate:
      allWins.length + allLosses.length > 0
        ? Math.round(
            (allWins.length / (allWins.length + allLosses.length)) * 10000
          ) / 100
        : 0,
    totalPnlPercent: Math.round(totalPnl * 100) / 100,
    totalPnlDollar: Math.round(totalPnlDollar * 100) / 100,
    avgWinPercent: Math.round(avgWin * 100) / 100,
    avgWinDollar: Math.round(avgWinDollar * 100) / 100,
    avgLossPercent: Math.round(avgLoss * 100) / 100,
    avgLossDollar: Math.round(avgLossDollar * 100) / 100,
    profitFactor:
      grossLoss > 0 ? Math.round((grossProfit / grossLoss) * 100) / 100 : 0,
    bestTrade:
      completed.length > 0
        ? completed.reduce((best, e) =>
            (e.pnlPercent || 0) > (best.pnlPercent || 0) ? e : best
          )
        : null,
    worstTrade:
      completed.length > 0
        ? completed.reduce((worst, e) =>
            (e.pnlPercent || 0) < (worst.pnlPercent || 0) ? e : worst
          )
        : null,
  };
}

/**
 * Get signal history, optionally filtered.
 */
export function getHistory({ assetType, symbol, limit = 50 } = {}) {
  let log = loadLog();
  if (assetType) log = log.filter((e) => e.assetType === assetType);
  if (symbol) log = log.filter((e) => e.symbol === symbol);
  return log.slice(-limit).reverse(); // newest first
}
