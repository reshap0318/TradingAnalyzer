import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const POSITIONS_FILE = path.join(__dirname, "../../data/positions.json");

/**
 * Position Tracker — Manual tracking of trades the user actually takes.
 * Supports both CRYPTO and SAHAM.
 */

function loadPositions() {
  try {
    if (fs.existsSync(POSITIONS_FILE)) {
      return JSON.parse(fs.readFileSync(POSITIONS_FILE, "utf8"));
    }
  } catch (e) {
    console.error("Error loading positions:", e.message);
  }
  return { open: [], closed: [] };
}

function savePositions(data) {
  const dir = path.dirname(POSITIONS_FILE);
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(POSITIONS_FILE, JSON.stringify(data, null, 2));
}

/**
 * Open a new position
 */
export function openPosition({
  symbol,
  assetType = "CRYPTO",
  side = "LONG",
  entryPrice,
  quantity,
  sl = null,
  tp = null,
  notes = "",
}) {
  const positions = loadPositions();

  // Check if already has open position for this symbol + assetType
  const existing = positions.open.find(
    (p) =>
      p.symbol === symbol && p.assetType === assetType && p.status === "OPEN"
  );
  if (existing) {
    return {
      success: false,
      error: `Already have open ${existing.side} position for ${symbol} @ ${existing.entryPrice}`,
    };
  }

  const position = {
    id: `${symbol}_${Date.now()}`,
    symbol,
    assetType,
    side,
    entryPrice,
    quantity,
    sl,
    tp,
    notes,
    openTime: new Date().toISOString(),
    status: "OPEN",
  };

  positions.open.push(position);
  savePositions(positions);
  console.log(
    `📍 Position opened: ${side} ${symbol} @ ${entryPrice} qty=${quantity}`
  );
  return { success: true, position };
}

/**
 * Close a position
 */
export function closePosition({
  symbol,
  assetType = null,
  exitPrice,
  reason = "MANUAL",
}) {
  const positions = loadPositions();

  const idx = positions.open.findIndex(
    (p) =>
      p.symbol === symbol &&
      p.status === "OPEN" &&
      (!assetType || p.assetType === assetType)
  );
  if (idx === -1) {
    return { success: false, error: `No open position found for ${symbol}` };
  }

  const position = positions.open[idx];
  const isBuy = position.side === "LONG";
  const pnlPerUnit = isBuy
    ? exitPrice - position.entryPrice
    : position.entryPrice - exitPrice;
  const pnlPercent = (pnlPerUnit / position.entryPrice) * 100;
  const pnlTotal = pnlPerUnit * position.quantity;
  const holdMs = Date.now() - new Date(position.openTime).getTime();
  const holdHours = Math.round((holdMs / (1000 * 60 * 60)) * 10) / 10;

  const closedPosition = {
    ...position,
    status: "CLOSED",
    exitPrice,
    exitTime: new Date().toISOString(),
    reason,
    pnlPerUnit: Math.round(pnlPerUnit * 100000000) / 100000000,
    pnlPercent: Math.round(pnlPercent * 100) / 100,
    pnlTotal: Math.round(pnlTotal * 100000000) / 100000000,
    holdHours,
  };

  // Move from open to closed
  positions.open.splice(idx, 1);
  positions.closed.push(closedPosition);
  savePositions(positions);

  const emoji = pnlPercent >= 0 ? "✅" : "❌";
  console.log(
    `${emoji} Position closed: ${symbol} @ ${exitPrice} | PnL: ${
      pnlPercent >= 0 ? "+" : ""
    }${closedPosition.pnlPercent}% (${reason})`
  );
  return { success: true, position: closedPosition };
}

/**
 * Get all open positions
 */
export function getOpenPositions(assetType = null) {
  const positions = loadPositions();
  if (assetType) {
    return positions.open.filter((p) => p.assetType === assetType);
  }
  return positions.open;
}

/**
 * Get closed positions history
 */
export function getPositionHistory({ assetType, symbol, limit = 50 } = {}) {
  const positions = loadPositions();
  let closed = positions.closed;
  if (assetType) closed = closed.filter((p) => p.assetType === assetType);
  if (symbol) closed = closed.filter((p) => p.symbol === symbol);
  return closed.slice(-limit).reverse(); // newest first
}

/**
 * Get position performance summary
 */
export function getPositionSummary(assetType = null) {
  const positions = loadPositions();
  let closed = positions.closed;
  if (assetType) closed = closed.filter((p) => p.assetType === assetType);

  const wins = closed.filter((p) => p.pnlPercent > 0);
  const losses = closed.filter((p) => p.pnlPercent <= 0);

  return {
    openPositions: assetType
      ? positions.open.filter((p) => p.assetType === assetType).length
      : positions.open.length,
    totalClosed: closed.length,
    wins: wins.length,
    losses: losses.length,
    winRate:
      closed.length > 0
        ? Math.round((wins.length / closed.length) * 10000) / 100
        : 0,
    totalPnlPercent:
      Math.round(closed.reduce((sum, p) => sum + p.pnlPercent, 0) * 100) / 100,
    avgWinPercent:
      wins.length > 0
        ? Math.round(
            (wins.reduce((sum, p) => sum + p.pnlPercent, 0) / wins.length) * 100
          ) / 100
        : 0,
    avgLossPercent:
      losses.length > 0
        ? Math.round(
            (losses.reduce((sum, p) => sum + p.pnlPercent, 0) / losses.length) *
              100
          ) / 100
        : 0,
  };
}
