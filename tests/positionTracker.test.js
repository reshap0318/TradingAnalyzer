import { describe, it, beforeEach, afterEach } from "node:test";
import assert from "node:assert/strict";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

import {
  openPosition,
  closePosition,
  getOpenPositions,
  getPositionHistory,
  getPositionSummary,
} from "../src/shared/positionTracker.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const POS_FILE = path.join(__dirname, "../data/positions.json");

let backup = null;

beforeEach(() => {
  if (fs.existsSync(POS_FILE)) {
    backup = fs.readFileSync(POS_FILE, "utf8");
  }
  fs.writeFileSync(POS_FILE, JSON.stringify({ open: [], closed: [] }));
});

afterEach(() => {
  if (backup !== null) {
    fs.writeFileSync(POS_FILE, backup);
  } else if (fs.existsSync(POS_FILE)) {
    fs.unlinkSync(POS_FILE);
  }
});

// ============================================================
// 1. openPosition
// ============================================================
describe("openPosition", () => {
  it("should open a LONG crypto position", () => {
    const result = openPosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      side: "LONG",
      entryPrice: 68000,
      quantity: 0.01,
      sl: 66000,
      tp: 70000,
    });
    assert.equal(result.success, true);
    assert.equal(result.position.symbol, "BTCUSDT");
    assert.equal(result.position.assetType, "CRYPTO");
    assert.equal(result.position.side, "LONG");
    assert.equal(result.position.status, "OPEN");
  });

  it("should open a LONG saham position", () => {
    const result = openPosition({
      symbol: "BBCA.JK",
      assetType: "SAHAM",
      side: "LONG",
      entryPrice: 8600,
      quantity: 500,
    });
    assert.equal(result.success, true);
    assert.equal(result.position.assetType, "SAHAM");
  });

  it("should reject duplicate position for same symbol + assetType", () => {
    openPosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      entryPrice: 68000,
      quantity: 0.01,
    });
    const r2 = openPosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      entryPrice: 69000,
      quantity: 0.02,
    });
    assert.equal(r2.success, false);
    assert.ok(r2.error.includes("Already have open"));
  });

  it("should allow same symbol with different assetType", () => {
    const r1 = openPosition({
      symbol: "TEST",
      assetType: "CRYPTO",
      entryPrice: 100,
      quantity: 1,
    });
    const r2 = openPosition({
      symbol: "TEST",
      assetType: "SAHAM",
      entryPrice: 100,
      quantity: 1,
    });
    assert.equal(r1.success, true);
    assert.equal(r2.success, true);
  });
});

// ============================================================
// 2. closePosition
// ============================================================
describe("closePosition", () => {
  it("should close position and calculate positive PnL", () => {
    openPosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      side: "LONG",
      entryPrice: 68000,
      quantity: 0.01,
    });
    const result = closePosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      exitPrice: 70000,
      reason: "TP",
    });
    assert.equal(result.success, true);
    assert.equal(result.position.status, "CLOSED");
    assert.ok(result.position.pnlPercent > 0);
    // (70000-68000)/68000 * 100 = 2.94%
    assert.equal(result.position.pnlPercent, 2.94);
    assert.equal(result.position.reason, "TP");
  });

  it("should close position and calculate negative PnL", () => {
    openPosition({
      symbol: "BBCA.JK",
      assetType: "SAHAM",
      side: "LONG",
      entryPrice: 8600,
      quantity: 500,
    });
    const result = closePosition({
      symbol: "BBCA.JK",
      assetType: "SAHAM",
      exitPrice: 8400,
      reason: "SL",
    });
    assert.equal(result.success, true);
    assert.ok(result.position.pnlPercent < 0);
    // pnlTotal = (8400-8600) * 500 = -100000
    assert.equal(result.position.pnlTotal, -100000);
  });

  it("should return error if no open position exists", () => {
    const result = closePosition({
      symbol: "NONEXIST",
      exitPrice: 1000,
    });
    assert.equal(result.success, false);
  });

  it("should not close crypto position when assetType is SAHAM", () => {
    openPosition({
      symbol: "TEST",
      assetType: "CRYPTO",
      entryPrice: 100,
      quantity: 1,
    });
    const result = closePosition({
      symbol: "TEST",
      assetType: "SAHAM",
      exitPrice: 110,
    });
    assert.equal(result.success, false);

    // Crypto position should still be open
    const open = getOpenPositions("CRYPTO");
    assert.equal(open.length, 1);
  });

  it("should calculate SHORT position PnL correctly", () => {
    openPosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      side: "SHORT",
      entryPrice: 68000,
      quantity: 0.01,
    });
    // SHORT: profit when price goes down
    const result = closePosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      exitPrice: 66000,
    });
    assert.ok(result.position.pnlPercent > 0); // profit
    // (68000-66000)/68000 * 100 = 2.94%
    assert.equal(result.position.pnlPercent, 2.94);
  });
});

// ============================================================
// 3. getOpenPositions — Asset Type Isolation
// ============================================================
describe("getOpenPositions — Isolation", () => {
  it("should only return positions of requested assetType", () => {
    openPosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      entryPrice: 68000,
      quantity: 0.01,
    });
    openPosition({
      symbol: "BBCA.JK",
      assetType: "SAHAM",
      entryPrice: 8600,
      quantity: 500,
    });

    const crypto = getOpenPositions("CRYPTO");
    const saham = getOpenPositions("SAHAM");

    assert.equal(crypto.length, 1);
    assert.equal(crypto[0].symbol, "BTCUSDT");

    assert.equal(saham.length, 1);
    assert.equal(saham[0].symbol, "BBCA.JK");
  });

  it("should return all positions when no assetType filter", () => {
    openPosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      entryPrice: 68000,
      quantity: 0.01,
    });
    openPosition({
      symbol: "BBCA.JK",
      assetType: "SAHAM",
      entryPrice: 8600,
      quantity: 500,
    });

    const all = getOpenPositions();
    assert.equal(all.length, 2);
  });
});

// ============================================================
// 4. getPositionSummary
// ============================================================
describe("getPositionSummary", () => {
  it("should calculate correct summary stats", () => {
    // Open & close 3 positions: 2 wins, 1 loss
    openPosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      side: "LONG",
      entryPrice: 68000,
      quantity: 0.01,
    });
    closePosition({
      symbol: "BTCUSDT",
      assetType: "CRYPTO",
      exitPrice: 70000,
      reason: "TP",
    });

    openPosition({
      symbol: "ETHUSDT",
      assetType: "CRYPTO",
      side: "LONG",
      entryPrice: 3000,
      quantity: 1,
    });
    closePosition({
      symbol: "ETHUSDT",
      assetType: "CRYPTO",
      exitPrice: 3300,
      reason: "TP",
    });

    openPosition({
      symbol: "SOLUSDT",
      assetType: "CRYPTO",
      side: "LONG",
      entryPrice: 100,
      quantity: 10,
    });
    closePosition({
      symbol: "SOLUSDT",
      assetType: "CRYPTO",
      exitPrice: 95,
      reason: "SL",
    });

    const summary = getPositionSummary("CRYPTO");
    assert.equal(summary.totalClosed, 3);
    assert.equal(summary.wins, 2);
    assert.equal(summary.losses, 1);
    assert.equal(summary.winRate, 66.67);
    assert.ok(summary.totalPnlPercent > 0);
    assert.ok(summary.avgWinPercent > 0);
    assert.ok(summary.avgLossPercent < 0);
  });

  it("should not include saham positions in crypto summary", () => {
    openPosition({
      symbol: "BBCA.JK",
      assetType: "SAHAM",
      entryPrice: 8600,
      quantity: 500,
    });
    closePosition({
      symbol: "BBCA.JK",
      assetType: "SAHAM",
      exitPrice: 9000,
    });

    const cryptoSummary = getPositionSummary("CRYPTO");
    assert.equal(cryptoSummary.totalClosed, 0);

    const sahamSummary = getPositionSummary("SAHAM");
    assert.equal(sahamSummary.totalClosed, 1);
    assert.equal(sahamSummary.wins, 1);
  });
});
