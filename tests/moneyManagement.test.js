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
    tp1: { price: 9200, percent: 6.98 },
    tp2: { price: 9800, percent: 13.95 },
    tp3: { price: 10500, percent: 22.09 },
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
    assert.ok(result.recommendation.lots > 0);
    assert.ok(result.recommendation.totalShares > 0);
    assert.ok(result.recommendation.positionValue > 0);
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
    assert.ok(weak.recommendation.lots < strong.recommendation.lots);
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
    assert.ok(result.analysis.riskPerShare > 0);
    assert.ok(result.analysis.riskPerSharePercent > 0);
    assert.ok(result.analysis.riskRewardRatio > 0);
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
    assert.ok(result.potentialProfit.atTP1 > 0);
    assert.ok(result.potentialProfit.atTP2 > result.potentialProfit.atTP1);
    assert.ok(result.potentialProfit.atTP3 > result.potentialProfit.atTP2);
  });

  it("should include trailing stop data", () => {
    const result = calculateMoneyManagement(
      defaultPortfolio,
      8600,
      8200,
      defaultTpsl,
      "BUY",
      80,
      "STOCK"
    );
    assert.ok(result.trailingStop.activationPrice > 8600);
    assert.ok(result.trailingStop.distance > 0);
  });
});

// ============================================================
// 2. calculateMoneyManagement — Negative scenarios
// ============================================================
describe("calculateMoneyManagement — Negative Scenarios", () => {
  it("should return isValid=false when signal is SELL", () => {
    const result = calculateMoneyManagement(
      { totalCapital: 100_000_000 },
      8600,
      8200,
      { tp1: { percent: 5 }, sl: { percent: -3 } },
      "SELL",
      50,
      "STOCK"
    );
    assert.equal(result.isValid, false);
    assert.ok(result.warnings.some((w) => w.includes("not BUY")));
  });

  it("should return error if no stop loss", () => {
    const result = calculateMoneyManagement(
      { totalCapital: 100_000_000 },
      8600,
      null, // No SL
      { tp1: { percent: 5 }, sl: { percent: -3 } },
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
        tp1: { price: 8700, percent: 1.16 },
        sl: { price: 8200, percent: -4.65 },
      },
      "BUY",
      50,
      "STOCK"
    );
    assert.equal(result.isValid, false);
    assert.ok(result.warnings.some((w) => w.includes("TP1")));
  });

  it("should warn for too many open positions", () => {
    const result = calculateMoneyManagement(
      { totalCapital: 100_000_000, currentPositions: 6 },
      8600,
      8200,
      {
        tp1: { price: 9000, percent: 4.65 },
        sl: { price: 8200, percent: -4.65 },
      },
      "BUY",
      50,
      "STOCK"
    );
    assert.ok(result.warnings.some((w) => w.includes("5 posisi")));
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
        tp1: { price: 72000, percent: 5.88 },
        sl: { price: 66000, percent: -2.94 },
      },
      "BUY",
      70,
      "CRYPTO"
    );
    assert.equal(result.isValid, true);
    // Crypto lots can be fractional
    assert.ok(typeof result.recommendation.lots === "number");
  });
});

// ============================================================
// 4. shouldTrade
// ============================================================
describe("shouldTrade", () => {
  it("should return true when ratio meets minimum", () => {
    const result = shouldTrade(
      { tp1: { percent: 6 }, sl: { percent: -3 } },
      1.5
    );
    assert.equal(result.trade, true);
    assert.ok(result.ratio >= 1.5);
  });

  it("should return false when ratio is too low", () => {
    const result = shouldTrade(
      { tp1: { percent: 2 }, sl: { percent: -3 } },
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
    const result = shouldTrade({ tp1: { percent: 5 } });
    assert.equal(result.trade, false);
  });

  it("should use custom minimum ratio", () => {
    const result = shouldTrade(
      { tp1: { percent: 3 }, sl: { percent: -2 } },
      2.0
    );
    assert.equal(result.trade, false); // 3/2 = 1.5 < 2.0
  });
});
