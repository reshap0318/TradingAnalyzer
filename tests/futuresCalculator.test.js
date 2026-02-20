import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { calculateFuturesPlan } from "../src/crypto/futuresCalculator.js";

// ============================================================
// 1. Basic LONG Calculation
// ============================================================
describe("calculateFuturesPlan — LONG", () => {
  const base = {
    capital: 1000,
    entryPrice: 68000,
    slPrice: 66000,
    side: "LONG",
    leverage: 10,
    tpsl: {
      tp: { price: 70000 },
    },
  };

  it("should calculate correct leverage and margin", () => {
    const result = calculateFuturesPlan(base);
    assert.equal(result.leverage, 10);
    assert.equal(result.side, "LONG");
    // margin = capital * MAX_POSITION_PERCENT (0.1) = 100
    assert.equal(result.margin.required, 100);
    // notional = 100 * 10 = 1000
    assert.equal(result.position.notionalValue, 1000);
  });

  it("should calculate liquidation price for LONG", () => {
    const result = calculateFuturesPlan(base);
    // LONG liqPrice = entry * (1 - 1/lev + maintenanceRate)
    // = 68000 * (1 - 0.1 + 0.005) = 68000 * 0.905 = 61540
    assert.ok(result.liquidation.price < base.entryPrice);
    assert.ok(result.liquidation.price > 0);
    assert.ok(result.liquidation.distancePercent > 0);
  });

  it("should calculate positive TP targets", () => {
    const result = calculateFuturesPlan(base);
    assert.ok(result.targets.tp.pnl > 0);

    // ROE should be leverage-amplified and positive
    assert.ok(result.targets.tp.roe > 0);
    // ROE = (diff/entry) * leverage * 100, should be roughly 29.41% for 10x
    assert.ok(result.targets.tp.roe > 20);
  });

  it("should calculate risk metrics", () => {
    const result = calculateFuturesPlan(base);
    // slPercent = (68000-66000)/68000 * 100 = 2.94%
    assert.ok(Math.abs(result.risk.slPercent - 2.94) < 0.1);
    assert.ok(result.risk.slLoss > 0);
    // effectiveRisk = slPercent * leverage = ~29.4% ROE
    assert.ok(result.risk.effectiveRisk.includes("ROE"));
  });

  it("should not warn about SL beyond liquidation for normal scenario", () => {
    const result = calculateFuturesPlan(base);
    assert.equal(result.liquidation.slBeyondLiquidation, false);
    const liqWarning = result.warnings.find((w) => w.includes("liquidat"));
    assert.equal(liqWarning, undefined);
  });

  it("should calculate fees", () => {
    const result = calculateFuturesPlan(base);
    assert.ok(result.fees.openFee > 0);
    assert.ok(result.fees.closeFee > 0);
    assert.ok(result.fees.totalFees > 0);
    assert.equal(
      result.fees.totalFees,
      result.fees.openFee + result.fees.closeFee
    );
    assert.ok(result.fees.fundingPer8h > 0);
  });
});

// ============================================================
// 2. SHORT Calculation
// ============================================================
describe("calculateFuturesPlan — SHORT", () => {
  it("should calculate liquidation price above entry for SHORT", () => {
    const result = calculateFuturesPlan({
      capital: 1000,
      entryPrice: 68000,
      slPrice: 70000,
      side: "SHORT",
      leverage: 10,
      tpsl: { tp: { price: 66000 } },
    });
    // SHORT liqPrice = entry * (1 + 1/lev - maintenanceRate)
    // = 68000 * (1 + 0.1 - 0.005) = 68000 * 1.095 = 74460
    assert.ok(result.liquidation.price > result.position.entryPrice);
  });

  it("should calculate positive PnL when SHORT is profitable", () => {
    const result = calculateFuturesPlan({
      capital: 1000,
      entryPrice: 68000,
      slPrice: 70000,
      side: "SHORT",
      leverage: 10,
      tpsl: {
        tp: { price: 66000 }, // Price going DOWN = profit for SHORT
      },
    });
    assert.ok(result.targets.tp.pnl > 0);
    assert.ok(result.targets.tp.roe > 0);
  });
});

// ============================================================
// 3. Warnings
// ============================================================
describe("calculateFuturesPlan — Warnings", () => {
  it("should warn when SL is beyond liquidation price (LONG)", () => {
    const result = calculateFuturesPlan({
      capital: 1000,
      entryPrice: 68000,
      slPrice: 50000, // Way below liquidation
      side: "LONG",
      leverage: 10,
      tpsl: {},
    });
    assert.equal(result.liquidation.slBeyondLiquidation, true);
    assert.ok(result.warnings.some((w) => w.includes("liquidat")));
  });

  it("should warn when SL is beyond liquidation price (SHORT)", () => {
    const result = calculateFuturesPlan({
      capital: 1000,
      entryPrice: 68000,
      slPrice: 90000, // Way above liquidation
      side: "SHORT",
      leverage: 10,
      tpsl: {},
    });
    assert.equal(result.liquidation.slBeyondLiquidation, true);
  });

  it("should warn for high leverage (>10x)", () => {
    const result = calculateFuturesPlan({
      capital: 1000,
      entryPrice: 68000,
      slPrice: 67000,
      side: "LONG",
      leverage: 15,
      tpsl: {},
    });
    assert.ok(result.warnings.some((w) => w.includes("15x")));
  });

  it("should cap leverage at MAX_LEVERAGE", () => {
    const result = calculateFuturesPlan({
      capital: 1000,
      entryPrice: 68000,
      slPrice: 67000,
      side: "LONG",
      leverage: 100, // Way above max (20)
      tpsl: {},
    });
    assert.ok(result.leverage <= 20);
  });
});

// ============================================================
// 4. Edge Cases
// ============================================================
describe("calculateFuturesPlan — Edge Cases", () => {
  it("should handle missing TP levels", () => {
    const result = calculateFuturesPlan({
      capital: 1000,
      entryPrice: 68000,
      slPrice: 66000,
      side: "LONG",
      leverage: 5,
      tpsl: {}, // No TP levels
    });
    assert.equal(result.targets.tp, null);
  });

  it("should use default leverage when not specified", () => {
    const result = calculateFuturesPlan({
      capital: 1000,
      entryPrice: 68000,
      slPrice: 66000,
      side: "LONG",
      tpsl: {},
    });
    // Default should be 5 from config
    assert.ok(result.leverage >= 1);
    assert.ok(result.leverage <= 20);
  });

  it("should handle very small entry price (low cap coins)", () => {
    const result = calculateFuturesPlan({
      capital: 100,
      entryPrice: 0.00001,
      slPrice: 0.000008,
      side: "LONG",
      leverage: 5,
      tpsl: { tp: { price: 0.000012 } },
    });
    assert.ok(result.position.quantity > 0);
    assert.ok(result.targets.tp.roe > 0);
  });
});
