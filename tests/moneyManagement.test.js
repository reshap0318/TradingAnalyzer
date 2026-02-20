import { describe, it } from "node:test";
import assert from "node:assert/strict";
import {
  calculateMoneyManagement,
  shouldTrade,
} from "../src/shared/moneyManagement.js";

// ============================================================
// 1. calculateMoneyManagement — Positive scenarios
// ============================================================
describe("calculateMoneyManagement — BUY Scenarios", () => {
  const defaultPortfolio = {
    totalCapital: 100_000_000, // 100M IDR
    maxLossPercent: 2,
    currentPositions: 0,
  };

  const defaultTpsl = {
    tp: { price: 9200, percent: 6.98 },
    sl: { price: 8200, percent: -4.65 },
  };

  it("should return valid recommendation for BUY signal", () => {
    const result = calculateMoneyManagement(
      defaultPortfolio,
      8600, // price
      8200, // slPrice
      defaultTpsl,
      "BUY",
      80, // trendStrength
      "STOCK"
    );
    assert.equal(result.isValid, true);
    assert.equal(result.signal, "BUY");
    assert.ok(result.totalLot > 0);
    assert.ok(result.priceLot.satuan > 0);
    assert.ok(result.priceLot.totalBelanja > 0);
  });

  it("should scale lots down for weak trend", () => {
    const strong = calculateMoneyManagement(
      defaultPortfolio,
      8600,
      8200,
      defaultTpsl,
      "BUY",
      80,
      "STOCK"
    );
    const weak = calculateMoneyManagement(
      defaultPortfolio,
      8600,
      8200,
      defaultTpsl,
      "BUY",
      30,
      "STOCK"
    );
    assert.ok(weak.totalLot < strong.totalLot);
  });

  it("should produce correct risk metrics", () => {
    const result = calculateMoneyManagement(
      defaultPortfolio,
      8600,
      8200,
      defaultTpsl,
      "BUY",
      80,
      "STOCK"
    );
    assert.ok(result.maksimalKerugian > 0);
    assert.ok(result.slPercent > 0);
    assert.ok(result.riskRewardRatio > 0);
  });

  it("should calculate potential profit at each TP", () => {
    const result = calculateMoneyManagement(
      defaultPortfolio,
      8600,
      8200,
      defaultTpsl,
      "BUY",
      80,
      "STOCK"
    );
    assert.ok(result.potensiKeuntungan > 0);
  });
});

// ============================================================
// 2. calculateMoneyManagement — Negative scenarios
// ============================================================
describe("calculateMoneyManagement — Negative Scenarios", () => {
  it("should return isValid=false when signal is SELL for STOCK", () => {
    const result = calculateMoneyManagement(
      { totalCapital: 100_000_000 },
      8600,
      8200,
      { tp: { percent: 5 }, sl: { percent: -3 } },
      "SELL",
      50,
      "STOCK"
    );
    assert.equal(result.isValid, false);
    assert.ok(result.warnings.some((w) => w.includes("not valid")));
  });

  it("should return error if no stop loss", () => {
    const result = calculateMoneyManagement(
      { totalCapital: 100_000_000 },
      8600,
      null, // No SL
      { tp: { percent: 5 }, sl: { percent: -3 } },
      "BUY",
      50,
      "STOCK"
    );
    assert.equal(result.isValid, false);
    assert.ok(result.reason.includes("Stop loss"));
  });

  it("should warn if TP1 < SL", () => {
    const result = calculateMoneyManagement(
      { totalCapital: 100_000_000 },
      8600,
      8200,
      {
        tp: { price: 8700, percent: 1.16 },
        sl: { price: 8200, percent: -4.65 },
      },
      "BUY",
      50,
      "STOCK"
    );
    assert.equal(result.isValid, false);
    assert.ok(result.warnings.some((w) => w.includes("TP")));
  });

  it("should warn for too many open positions", () => {
    const result = calculateMoneyManagement(
      { totalCapital: 100_000_000, currentPositions: 6 },
      8600,
      8200,
      {
        tp: { price: 9000, percent: 4.65 },
        sl: { price: 8200, percent: -4.65 },
      },
      "BUY",
      50,
      "STOCK"
    );
    assert.ok(result.warnings.some((w) => w.includes("maksimal posisi")));
  });
});

// ============================================================
// 3. calculateMoneyManagement — Crypto
// ============================================================
describe("calculateMoneyManagement — Crypto", () => {
  it("should handle fractional lot sizes for crypto", () => {
    const result = calculateMoneyManagement(
      { totalCapital: 1000 },
      68000,
      66000,
      {
        tp: { price: 72000, percent: 5.88 },
        sl: { price: 66000, percent: -2.94 },
      },
      "BUY",
      70,
      "CRYPTO"
    );
    assert.equal(result.isValid, true);
    // Crypto lots can be fractional
    assert.ok(typeof result.totalLot === "number");
  });
});

// ============================================================
// 4. shouldTrade
// ============================================================
describe("shouldTrade", () => {
  it("should return true when ratio meets minimum", () => {
    const result = shouldTrade(
      { tp: { percent: 6 }, sl: { percent: -3 } },
      1.5
    );
    assert.equal(result.trade, true);
    assert.ok(result.ratio >= 1.5);
  });

  it("should return false when ratio is too low", () => {
    const result = shouldTrade(
      { tp: { percent: 2 }, sl: { percent: -3 } },
      1.5
    );
    assert.equal(result.trade, false);
    assert.ok(result.reason.includes("Risk/Reward"));
  });

  it("should return false if TP/SL not available", () => {
    const result = shouldTrade(null);
    assert.equal(result.trade, false);
  });

  it("should return false if sl is missing", () => {
    const result = shouldTrade({ tp: { percent: 5 } });
    assert.equal(result.trade, false);
  });

  it("should use custom minimum ratio", () => {
    const result = shouldTrade(
      { tp: { percent: 3 }, sl: { percent: -2 } },
      2.0
    );
    assert.equal(result.trade, false); // 3/2 = 1.5 < 2.0
  });
});
