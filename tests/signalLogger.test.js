import { describe, it, beforeEach, afterEach } from "node:test";
import assert from "node:assert/strict";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

// --- Setup: override signalLogger's data file with a temp file ---
const __dirname = path.dirname(fileURLToPath(import.meta.url));
const TEMP_LOG = path.join(__dirname, "../data/signal_log_test.json");

// We need to patch the module's file path. Since signalLogger uses a hardcoded
// path, we'll create a wrapper that loads/saves to the same file but we clean
// it before/after each test.

// Import the actual functions
import {
  logSignal,
  closePendingSignal,
  updateOutcomes,
  getSummary,
  getHistory,
  getActive,
} from "../src/shared/signalLogger.js";

const ACTIVE_FILE = path.join(__dirname, "../data/active_trade.json");
const HISTORY_FILE = path.join(__dirname, "../data/history.json");
const SUMMARY_FILE = path.join(__dirname, "../data/summary.json");

function cleanLogs() {
  if (!fs.existsSync(path.dirname(ACTIVE_FILE)))
    fs.mkdirSync(path.dirname(ACTIVE_FILE), { recursive: true });
  fs.writeFileSync(ACTIVE_FILE, "[]");
  fs.writeFileSync(HISTORY_FILE, "[]");
  fs.writeFileSync(SUMMARY_FILE, "{}");
}

let backups = {};

beforeEach(() => {
  backups = {
    active: fs.existsSync(ACTIVE_FILE)
      ? fs.readFileSync(ACTIVE_FILE, "utf8")
      : null,
    history: fs.existsSync(HISTORY_FILE)
      ? fs.readFileSync(HISTORY_FILE, "utf8")
      : null,
    summary: fs.existsSync(SUMMARY_FILE)
      ? fs.readFileSync(SUMMARY_FILE, "utf8")
      : null,
  };
  cleanLogs();
});

afterEach(() => {
  if (backups.active !== null) fs.writeFileSync(ACTIVE_FILE, backups.active);
  else if (fs.existsSync(ACTIVE_FILE)) fs.unlinkSync(ACTIVE_FILE);
  if (backups.history !== null) fs.writeFileSync(HISTORY_FILE, backups.history);
  else if (fs.existsSync(HISTORY_FILE)) fs.unlinkSync(HISTORY_FILE);
  if (backups.summary !== null) fs.writeFileSync(SUMMARY_FILE, backups.summary);
  else if (fs.existsSync(SUMMARY_FILE)) fs.unlinkSync(SUMMARY_FILE);
});

// ============================================================
// Helper: create a dummy signal
// ============================================================
function dummySignal(overrides = {}) {
  return {
    symbol: "BTCUSDT",
    assetType: "CRYPTO",
    signal: "BUY",
    entryPrice: 68000,
    confidence: 72,
    score: 65,
    strength: "STRONG",
    tp: { price: 70000 },
    sl: { price: 66000 },
    riskReward: { tp: 1.5 },
    timeframeAlignment: "BULLISH_ALIGNED",
    marketTrend: "BULLISH",
    ...overrides,
  };
}

// ============================================================
// 1. logSignal — Basic Logging
// ============================================================
describe("logSignal", () => {
  it("should log a new BUY signal", () => {
    const result = logSignal(dummySignal());
    assert.equal(result.logged, true);
    assert.ok(result.id);

    const history = getActive({ assetType: "CRYPTO" });
    assert.equal(history.length, 1);
    assert.equal(history[0].signal, "BUY");
    assert.equal(history[0].outcome, "PENDING");
    assert.equal(history[0].entryPrice, 68000);
  });

  it("should log a SELL signal for crypto", () => {
    const result = logSignal(dummySignal({ signal: "SELL" }));
    assert.equal(result.logged, true);

    const history = getActive({ assetType: "CRYPTO" });
    assert.equal(history[0].signal, "SELL");
  });

  it("should correctly store TP/SL values", () => {
    logSignal(dummySignal());
    const history = getActive({ assetType: "CRYPTO" });
    assert.equal(history[0].tp, 70000);
    assert.equal(history[0].sl, 66000);
  });
});

// ============================================================
// 2. logSignal — Dedup (1 PENDING per symbol)
// ============================================================
describe("logSignal — Dedup", () => {
  it("should skip duplicate BUY for same symbol", () => {
    const r1 = logSignal(dummySignal());
    assert.equal(r1.logged, true);

    const r2 = logSignal(dummySignal({ entryPrice: 69000 }));
    assert.equal(r2.logged, false);
    assert.ok(r2.reason.includes("Already tracking"));

    // Still only 1 entry
    const history = getActive({ assetType: "CRYPTO" });
    assert.equal(history.length, 1);
    assert.equal(history[0].entryPrice, 68000); // Original price kept
  });

  it("should allow same direction for different symbols", () => {
    logSignal(dummySignal({ symbol: "BTCUSDT" }));
    const r2 = logSignal(dummySignal({ symbol: "ETHUSDT", entryPrice: 3400 }));
    assert.equal(r2.logged, true);

    const history = getActive({ assetType: "CRYPTO" });
    assert.equal(history.length, 2);
  });
});

// ============================================================
// 3. logSignal — SIGNAL_REVERSED (crypto)
// ============================================================
describe("logSignal — SIGNAL_REVERSED (Crypto)", () => {
  it("should reverse BUY→SELL and create new SELL entry", () => {
    // Log BUY first
    logSignal(dummySignal({ signal: "BUY", entryPrice: 68000 }));

    // Now log SELL — should reverse BUY and create SELL
    const r2 = logSignal(dummySignal({ signal: "SELL", entryPrice: 67000 }));
    assert.equal(r2.logged, true);

    const history = getHistory({ assetType: "CRYPTO" });
    const active = getActive({ assetType: "CRYPTO" });
    assert.equal(history.length, 1);
    assert.equal(active.length, 1);

    // active has the SELL
    assert.equal(active[0].signal, "SELL");
    assert.equal(active[0].outcome, "PENDING");

    // history has the reversed BUY
    assert.equal(history[0].signal, "BUY");
    assert.equal(history[0].outcome, "SIGNAL_REVERSED");
    assert.equal(history[0].exitPrice, 67000);
    // BUY @ 68000, exit @ 67000 → -1.47%
    assert.ok(history[0].pnlPercent < 0);
  });

  it("should reverse SELL→BUY and create new BUY entry", () => {
    logSignal(dummySignal({ signal: "SELL", entryPrice: 68000 }));
    logSignal(dummySignal({ signal: "BUY", entryPrice: 66000 }));

    const history = getHistory({ assetType: "CRYPTO" });
    // Reversed SELL should be profitable (SELL @68000, exit @66000)
    const reversed = history.find((h) => h.outcome === "SIGNAL_REVERSED");
    assert.ok(reversed.pnlPercent > 0);
  });

  it("should calculate PnL correctly on reverse", () => {
    logSignal(dummySignal({ signal: "BUY", entryPrice: 50000 }));
    logSignal(dummySignal({ signal: "SELL", entryPrice: 55000 }));

    const history = getHistory({ assetType: "CRYPTO" });
    const reversed = history.find((h) => h.outcome === "SIGNAL_REVERSED");
    // BUY @50000, exit @55000 → +10%
    assert.equal(reversed.pnlPercent, 10);
  });
});

// ============================================================
// 4. closePendingSignal (saham behavior)
// ============================================================
describe("closePendingSignal — Saham Behavior", () => {
  it("should close PENDING BUY without creating new entry", () => {
    logSignal(
      dummySignal({
        symbol: "BBCA.JK",
        assetType: "SAHAM",
        signal: "BUY",
        entryPrice: 8600,
      })
    );

    const result = closePendingSignal("BBCA.JK", 8700);
    assert.equal(result.closed, true);
    assert.ok(result.pnlPercent > 0); // profit

    const history = getHistory({ assetType: "SAHAM" });
    assert.equal(history.length, 1); // No new entry created
    assert.equal(history[0].outcome, "SIGNAL_REVERSED");
    assert.equal(history[0].exitPrice, 8700);
  });

  it("should return closed=false if no PENDING exists", () => {
    const result = closePendingSignal("NONEXIST", 1000);
    assert.equal(result.closed, false);
  });

  it("should calculate negative PnL when price dropped", () => {
    logSignal(
      dummySignal({
        symbol: "BBCA.JK",
        assetType: "SAHAM",
        signal: "BUY",
        entryPrice: 8600,
      })
    );

    const result = closePendingSignal("BBCA.JK", 8400);
    assert.ok(result.pnlPercent < 0); // loss
  });
});

// ============================================================
// 5. Asset Type Isolation
// ============================================================
describe("Asset Type Isolation", () => {
  it("should separate SAHAM and CRYPTO signals", () => {
    logSignal(
      dummySignal({
        symbol: "BBCA.JK",
        assetType: "SAHAM",
        signal: "BUY",
        entryPrice: 8600,
      })
    );
    logSignal(
      dummySignal({
        symbol: "BTCUSDT",
        assetType: "CRYPTO",
        signal: "BUY",
        entryPrice: 68000,
      })
    );

    const sahamHistory = getActive({ assetType: "SAHAM" });
    const cryptoHistory = getActive({ assetType: "CRYPTO" });

    assert.equal(sahamHistory.length, 1);
    assert.equal(sahamHistory[0].symbol, "BBCA.JK");

    assert.equal(cryptoHistory.length, 1);
    assert.equal(cryptoHistory[0].symbol, "BTCUSDT");
  });

  it("should return separate summaries", () => {
    logSignal(
      dummySignal({
        symbol: "BBCA.JK",
        assetType: "SAHAM",
        signal: "BUY",
        entryPrice: 8600,
      })
    );
    logSignal(
      dummySignal({
        symbol: "BTCUSDT",
        assetType: "CRYPTO",
        signal: "BUY",
        entryPrice: 68000,
      })
    );

    const sahamSummary = getSummary("SAHAM");
    const cryptoSummary = getSummary("CRYPTO");

    assert.equal(sahamSummary.totalSignals, 1);
    assert.equal(cryptoSummary.totalSignals, 1);
  });
});

// ============================================================
// 6. updateOutcomes — TP/SL/Expire
// ============================================================
describe("updateOutcomes", () => {
  it("should mark TP hit when price reaches target", async () => {
    logSignal(dummySignal({ signal: "BUY", entryPrice: 68000 }));
    // tp = 70000

    // Mock price function that returns 70500 (above TP)
    await updateOutcomes(async () => 70500, "CRYPTO");

    const history = getHistory({ assetType: "CRYPTO" });
    assert.equal(history[0].outcome, "TP_HIT");
  });

  it("should mark SL hit when price drops to stop loss", async () => {
    logSignal(dummySignal({ signal: "BUY", entryPrice: 68000 }));
    // sl = 66000

    await updateOutcomes(async () => 65500, "CRYPTO");

    const history = getHistory({ assetType: "CRYPTO" });
    assert.equal(history[0].outcome, "SL_HIT");
    assert.ok(history[0].pnlPercent < 0);
  });

  it("should keep PENDING when price is between SL and TP", async () => {
    logSignal(
      dummySignal({
        signal: "BUY",
        entryPrice: 68000,
        sl: { price: 65000 },
        tp: { price: 72000 },
      })
    );

    // Price at 69000 — above SL, below TP → should stay PENDING
    await updateOutcomes(async () => 69000, "CRYPTO");

    const history = getActive({ assetType: "CRYPTO" });
    assert.equal(history[0].outcome, "PENDING");
  });

  it("should not update signals from wrong asset type", async () => {
    logSignal(
      dummySignal({
        symbol: "BBCA.JK",
        assetType: "SAHAM",
        signal: "BUY",
        entryPrice: 8600,
        tp: { price: 8800 },
      })
    );

    // Update only CRYPTO — saham should remain PENDING
    await updateOutcomes(async () => 9000, "CRYPTO");

    const history = getActive({ assetType: "SAHAM" });
    assert.equal(history[0].outcome, "PENDING");
  });

  it("should skip signals when getPriceFn throws", async () => {
    logSignal(dummySignal({ signal: "BUY", entryPrice: 68000 }));

    await updateOutcomes(async () => {
      throw new Error("API error");
    }, "CRYPTO");

    const history = getActive({ assetType: "CRYPTO" });
    assert.equal(history[0].outcome, "PENDING"); // Unchanged
  });

  it("should handle SELL signals correctly (SL/TP reversed)", async () => {
    logSignal(
      dummySignal({
        signal: "SELL",
        entryPrice: 68000,
        tp: { price: 66000 }, // For SELL, price must go DOWN to hit TP
        sl: { price: 70000 }, // For SELL, price going UP hits SL
      })
    );

    // Price goes up to 70500 → SL hit for SELL
    await updateOutcomes(async () => 70500, "CRYPTO");

    const history = getHistory({ assetType: "CRYPTO" });
    assert.equal(history[0].outcome, "SL_HIT");
  });
});

// ============================================================
// 7. getSummary — Statistics
// ============================================================
describe("getSummary", () => {
  it("should return correct stats with mixed outcomes", () => {
    // Create some completed signals by manipulating log directly
    const log = [
      {
        id: "TEST_1",
        symbol: "BTCUSDT",
        assetType: "CRYPTO",
        signal: "BUY",
        entryPrice: 60000,
        outcome: "TP_HIT",
        pnlPercent: 3.0,
      },
      {
        id: "TEST_2",
        symbol: "ETHUSDT",
        assetType: "CRYPTO",
        signal: "BUY",
        entryPrice: 3000,
        outcome: "TP_HIT",
        pnlPercent: 5.0,
      },
      {
        id: "TEST_3",
        symbol: "BTCUSDT",
        assetType: "CRYPTO",
        signal: "SELL",
        entryPrice: 70000,
        outcome: "SL_HIT",
        pnlPercent: -2.0,
      },
      {
        id: "TEST_4",
        symbol: "SOLUSDT",
        assetType: "CRYPTO",
        signal: "BUY",
        entryPrice: 100,
        outcome: "SIGNAL_REVERSED",
        pnlPercent: 1.5,
      },
      {
        id: "TEST_5",
        symbol: "DOGEUSDT",
        assetType: "CRYPTO",
        signal: "BUY",
        entryPrice: 0.08,
        outcome: "EXPIRED",
        pnlPercent: 0,
      },
      {
        id: "TEST_6",
        symbol: "BTCUSDT",
        assetType: "CRYPTO",
        signal: "BUY",
        entryPrice: 68000,
        outcome: "PENDING",
      },
    ];
    fs.writeFileSync(
      HISTORY_FILE,
      JSON.stringify(
        log.filter((l) => l.outcome !== "PENDING"),
        null,
        2
      )
    );
    fs.writeFileSync(
      ACTIVE_FILE,
      JSON.stringify(
        log.filter((l) => l.outcome === "PENDING"),
        null,
        2
      )
    );

    const summary = getSummary("CRYPTO");
    assert.equal(summary.totalSignals, 6);
    assert.equal(summary.pending, 1);
    assert.equal(summary.completed, 5);
    // Wins: TP_HIT, TP_HIT, SIGNAL_REVERSED (+1.5%) = 3
    assert.equal(summary.wins, 3);
    // Losses: SL_HIT = 1
    assert.equal(summary.losses, 1);
    assert.equal(summary.reversed, 1);
    assert.equal(summary.expired, 1);
    // winRate = 3 / (3+1) = 75%
    assert.equal(summary.winRate, 75);
    // totalPnl = 3 + 5 + (-2) + 1.5 + 0 = 7.5
    assert.equal(summary.totalPnlPercent, 7.5);
  });

  it("should return zeros for empty log", () => {
    const summary = getSummary("CRYPTO");
    assert.equal(summary.totalSignals, 0);
    assert.equal(summary.winRate, 0);
    assert.equal(summary.profitFactor, 0);
  });
});
