import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ACTIVE_FILE = path.join(__dirname, "../../data/active_trade.json");
const HISTORY_FILE = path.join(__dirname, "../../data/history.json");
const SUMMARY_FILE = path.join(__dirname, "../../data/summary.json");

/**
 * Signal Logger — Automatically records every BUY/SELL signal for performance evaluation.
 * Separated into Active, History, and Summary logs.
 */

function loadJson(file, defaultVal = []) {
  try {
    if (fs.existsSync(file)) {
      return JSON.parse(fs.readFileSync(file, "utf8"));
    }
  } catch (e) {
    console.error(`Error loading log file ${file}:`, e.message);
  }
  return defaultVal;
}

function saveJson(file, data) {
  const dir = path.dirname(file);
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(file, JSON.stringify(data, null, 2));
}

function getWIBTimestamp(date = new Date()) {
  return (
    date
      .toLocaleString("sv-SE", { timeZone: "Asia/Jakarta" })
      .replace(" ", "T") + "+07:00"
  );
}

function moveToHistory(entry) {
  const history = loadJson(HISTORY_FILE, []);
  history.push(entry);
  saveJson(HISTORY_FILE, history);
}

function updateSummaryData() {
  const summaryCrypto = calculateSummaryStats("CRYPTO");
  const summarySaham = calculateSummaryStats("SAHAM");

  const summary = {
    lastUpdated: getWIBTimestamp(),
    CRYPTO: summaryCrypto,
    SAHAM: summarySaham,
  };

  saveJson(SUMMARY_FILE, summary);
  return summary;
}

function calculateSummaryStats(assetType) {
  const activeLogs = loadJson(ACTIVE_FILE, []).filter(
    (e) => e.assetType === assetType
  );
  const historyLogs = loadJson(HISTORY_FILE, []).filter(
    (e) => e.assetType === assetType
  );

  const completed = historyLogs.filter((e) => e.outcome !== "PENDING");
  const wins = completed.filter((e) => ["TP_HIT"].includes(e.outcome));
  const losses = completed.filter((e) => e.outcome === "SL_HIT");
  const reversed = completed.filter((e) => e.outcome === "SIGNAL_REVERSED");
  const reversedWins = reversed.filter((e) => (e.pnlPercent || 0) > 0);
  const reversedLosses = reversed.filter((e) => (e.pnlPercent || 0) <= 0);
  const expired = completed.filter((e) => e.outcome === "EXPIRED");
  const pending = activeLogs;

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
    totalSignals: pending.length + completed.length,
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

export function logSignal({
  symbol,
  assetType = "CRYPTO",
  signal,
  entryPrice,
  confidence,
  score,
  strength,
  tp,
  sl,
  riskReward,
  timeframeAlignment,
  marketTrend,
  allocatedAmount = 0,
}) {
  const active = loadJson(ACTIVE_FILE, []);
  const now = new Date();

  const pendingIdx = active.findIndex(
    (entry) => entry.symbol === symbol && entry.outcome === "PENDING"
  );

  if (pendingIdx !== -1) {
    const pending = active[pendingIdx];

    if (pending.signal === signal) {
      return {
        logged: false,
        reason: `Already tracking ${signal} for ${symbol} since ${pending.timestamp}`,
      };
    }

    const isBuy = pending.signal === "BUY";
    const pnl = isBuy
      ? ((entryPrice - pending.entryPrice) / pending.entryPrice) * 100
      : ((pending.entryPrice - entryPrice) / pending.entryPrice) * 100;
    const ageHours = (now - new Date(pending.timestamp)) / (1000 * 60 * 60);

    pending.outcome = "SIGNAL_REVERSED";
    pending.exitPrice = entryPrice;
    pending.exitTime = getWIBTimestamp(now);
    pending.pnlPercent = Math.round(pnl * 100) / 100;
    pending.pnlDollar =
      Math.round((pending.allocatedAmount || 0) * (pnl / 100) * 100) / 100;
    pending.holdHours = Math.round(ageHours * 10) / 10;

    const emoji = pnl >= 0 ? "✅" : "❌";
    console.log(
      `🔄 Signal reversed: ${pending.signal}→${signal} ${symbol} | PnL: ${
        pnl >= 0 ? "+" : ""
      }${pending.pnlPercent}%`
    );

    active.splice(pendingIdx, 1);
    moveToHistory(pending);
    updateSummaryData();
  }

  const entry = {
    id: `${symbol}_${now.getTime()}`,
    symbol,
    assetType,
    signal,
    timestamp: getWIBTimestamp(now),
    entryPrice,
    confidence,
    score,
    strength,
    tp: tp?.price || null,
    sl: sl?.price || null,
    riskReward: riskReward?.tp || null,
    timeframeAlignment,
    marketTrend,
    allocatedAmount: Math.round(allocatedAmount * 10000) / 10000,
    outcome: "PENDING",
    highestPrice: entryPrice,
    lowestPrice: entryPrice,
    exitPrice: null,
    exitTime: null,
    pnlPercent: null,
    pnlDollar: null,
    holdHours: null,
  };

  active.push(entry);
  saveJson(ACTIVE_FILE, active);
  updateSummaryData();
  console.log(`📝 Signal logged: ${symbol} ${signal} @ ${entryPrice}`);
  return { logged: true, id: entry.id };
}

export function closePendingSignal(symbol, currentPrice) {
  const active = loadJson(ACTIVE_FILE, []);
  const now = new Date();

  const pendingIdx = active.findIndex(
    (entry) => entry.symbol === symbol && entry.outcome === "PENDING"
  );
  if (pendingIdx === -1)
    return { closed: false, reason: "No PENDING signal found" };

  const pending = active[pendingIdx];
  const isBuy = pending.signal === "BUY";
  const pnl = isBuy
    ? ((currentPrice - pending.entryPrice) / pending.entryPrice) * 100
    : ((pending.entryPrice - currentPrice) / pending.entryPrice) * 100;
  const ageHours = (now - new Date(pending.timestamp)) / (1000 * 60 * 60);

  pending.outcome = "SIGNAL_REVERSED";
  pending.exitPrice = currentPrice;
  pending.exitTime = getWIBTimestamp(now);
  pending.pnlPercent = Math.round(pnl * 100) / 100;
  pending.pnlDollar =
    Math.round((pending.allocatedAmount || 0) * (pnl / 100) * 100) / 100;
  pending.holdHours = Math.round(ageHours * 10) / 10;

  console.log(
    `🔄 Signal closed (reversed): ${
      pending.signal
    } ${symbol} @ ${currentPrice} | PnL: ${pnl >= 0 ? "+" : ""}${
      pending.pnlPercent
    }%`
  );

  active.splice(pendingIdx, 1);
  moveToHistory(pending);
  updateSummaryData();
  saveJson(ACTIVE_FILE, active);

  return { closed: true, pnlPercent: pending.pnlPercent };
}

export async function updateOutcomes(getPriceFn, assetType = null) {
  const active = loadJson(ACTIVE_FILE, []);
  let updated = 0;

  // Track the actual current state of active array by iterating backwards to allow splicing
  for (let i = active.length - 1; i >= 0; i--) {
    const entry = active[i];
    if (entry.outcome !== "PENDING") continue;
    if (assetType && entry.assetType !== assetType) continue;

    const timestampMatch = entry.timestamp.match(
      /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})/
    );
    let entryDate;
    if (timestampMatch) {
      // Create date assuming local config (sv-SE formats is easily parsable but wait, the WIB timestamp ends in +07:00, which IS standard ISO, so new Date(entry.timestamp) works natively!)
      entryDate = new Date(entry.timestamp);
    } else {
      entryDate = new Date(); // fallback
    }

    const ageHours = (Date.now() - entryDate.getTime()) / (1000 * 60 * 60);
    const maxAge = entry.assetType === "CRYPTO" ? 48 : 240;

    if (ageHours > maxAge) {
      entry.outcome = "EXPIRED";
      entry.holdHours = Math.round(ageHours);
      entry.pnlPercent = 0;
      entry.pnlDollar = 0;

      const removed = active.splice(i, 1)[0];
      moveToHistory(removed);
      updated++;
      continue;
    }

    let currentPrice;
    try {
      currentPrice = await getPriceFn(entry.symbol, entry.assetType);
      if (!currentPrice) continue;
    } catch {
      continue;
    }

    entry.highestPrice = Math.max(
      entry.highestPrice || entry.entryPrice,
      currentPrice
    );
    entry.lowestPrice = Math.min(
      entry.lowestPrice || entry.entryPrice,
      currentPrice
    );

    const isBuy = entry.signal === "BUY";
    let isHit = false;

    if (entry.sl) {
      const slHit = isBuy ? currentPrice <= entry.sl : currentPrice >= entry.sl;
      if (slHit) {
        entry.outcome = "SL_HIT";
        entry.exitPrice = entry.sl;
        entry.exitTime = getWIBTimestamp();
        entry.pnlPercent = isBuy
          ? ((entry.sl - entry.entryPrice) / entry.entryPrice) * 100
          : ((entry.entryPrice - entry.sl) / entry.entryPrice) * 100;
        entry.pnlDollar =
          Math.round(
            (entry.allocatedAmount || 0) * (entry.pnlPercent / 100) * 100
          ) / 100;
        entry.holdHours = Math.round(ageHours);

        const removed = active.splice(i, 1)[0];
        moveToHistory(removed);
        updated++;
        isHit = true;
      }
    }

    if (!isHit) {
      const tpChecks = [{ price: entry.tp, label: "TP_HIT" }];

      for (const tp of tpChecks) {
        if (!tp.price) continue;
        const tpHit = isBuy
          ? currentPrice >= tp.price
          : currentPrice <= tp.price;
        if (tpHit) {
          entry.outcome = tp.label;
          entry.exitPrice = tp.price;
          entry.exitTime = getWIBTimestamp();
          entry.pnlPercent = isBuy
            ? ((tp.price - entry.entryPrice) / entry.entryPrice) * 100
            : ((entry.entryPrice - tp.price) / entry.entryPrice) * 100;
          entry.pnlDollar =
            Math.round(
              (entry.allocatedAmount || 0) * (entry.pnlPercent / 100) * 100
            ) / 100;
          entry.holdHours = Math.round(ageHours);

          const removed = active.splice(i, 1)[0];
          moveToHistory(removed);
          updated++;
          break;
        }
      }
    }
  }

  // Check if we still need to save active logs because we modified highestPrice/lowestPrice
  saveJson(ACTIVE_FILE, active);

  if (updated > 0) {
    updateSummaryData();
    console.log(`📊 Signal outcomes updated: ${updated} entries`);
  }
  return updated;
}

export function getCapitalStatus(assetType, initialCapital) {
  const active = loadJson(ACTIVE_FILE, []).filter(
    (e) => e.assetType === assetType
  );
  const summary = loadJson(SUMMARY_FILE, {})[assetType] || {
    totalPnlDollar: 0,
  };

  const allocated = active.reduce(
    (sum, e) => sum + (e.allocatedAmount || 0),
    0
  );
  const realizedPnl = summary.totalPnlDollar || 0;
  const available = initialCapital - allocated + realizedPnl;

  return {
    initialCapital: Math.round(initialCapital * 100) / 100,
    allocated: Math.round(allocated * 100) / 100,
    realizedPnl: Math.round(realizedPnl * 100) / 100,
    available: Math.round(available * 100) / 100,
    openPositions: active.length,
  };
}

export function getSummary(assetType = null) {
  const summaryAll = updateSummaryData();
  return assetType ? summaryAll[assetType] : summaryAll;
}

export function getHistory({ assetType, symbol, limit = 50 } = {}) {
  let log = loadJson(HISTORY_FILE, []);
  if (assetType) log = log.filter((e) => e.assetType === assetType);
  if (symbol) log = log.filter((e) => e.symbol === symbol);
  return log.slice(-limit).reverse();
}
