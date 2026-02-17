import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { findSupportResistance } from "../src/shared/supportResistance.js";
import { calculateTPSL } from "../src/shared/tpslCalculator.js";

// ============================================================
// Helper: Generate OHLC data with known S/R levels
// ============================================================
function generateOHLC(count = 200, basePrice = 100, amp = 10) {
  const data = [];
  for (let i = 0; i < count; i++) {
    const sine = Math.sin(i / 20) * amp; // Creates oscillation = natural S/R
    const center = basePrice + (i / count) * 10;
    const close = center + sine + (Math.random() - 0.5) * 2;
    const open = close + (Math.random() - 0.5) * 3;
    const high = Math.max(open, close) + Math.random() * 2;
    const low = Math.min(open, close) - Math.random() * 2;
    data.push({ open, high, low, close, volume: 1000000 });
  }
  return data;
}

// ============================================================
// 1. Support/Resistance
// ============================================================
describe("findSupportResistance", () => {
  it("should find supports and resistances from OHLC data", () => {
    const data = generateOHLC(200);
    const sr = findSupportResistance(data);

    assert.ok(Array.isArray(sr.supports));
    assert.ok(Array.isArray(sr.resistances));
    assert.ok(typeof sr.currentPrice === "number");
  });

  it("should have supports below current price", () => {
    const data = generateOHLC(200);
    const sr = findSupportResistance(data);
    sr.supports.forEach((s) => {
      assert.ok(
        s.level < sr.currentPrice,
        `Support ${s.level} should be below current ${sr.currentPrice}`
      );
    });
  });

  it("should have resistances above current price", () => {
    const data = generateOHLC(200);
    const sr = findSupportResistance(data);
    sr.resistances.forEach((r) => {
      assert.ok(
        r.level > sr.currentPrice,
        `Resistance ${r.level} should be above current ${sr.currentPrice}`
      );
    });
  });

  it("should return at most 3 supports and 3 resistances", () => {
    const data = generateOHLC(300);
    const sr = findSupportResistance(data);
    assert.ok(sr.supports.length <= 3);
    assert.ok(sr.resistances.length <= 3);
  });

  it("should return empty arrays for insufficient data", () => {
    const sr = findSupportResistance([{ open: 1, high: 2, low: 0, close: 1 }]);
    assert.equal(sr.supports.length, 0);
    assert.equal(sr.resistances.length, 0);
  });

  it("should return empty for null input", () => {
    const sr = findSupportResistance(null);
    assert.equal(sr.supports.length, 0);
    assert.equal(sr.resistances.length, 0);
  });

  it("should include pivot type info", () => {
    const data = generateOHLC(300);
    const sr = findSupportResistance(data);
    sr.supports.forEach((s) => assert.ok(s.type));
    sr.resistances.forEach((r) => assert.ok(r.type));
  });
});

// ============================================================
// 2. TP/SL Calculator — BUY (LONG)
// ============================================================
describe("calculateTPSL — BUY", () => {
  const data = generateOHLC(200, 100);
  const price = data[data.length - 1].close;

  it("should calculate TP1, TP2, TP3 above entry for BUY", () => {
    const result = calculateTPSL(data, "BUY", price);
    assert.ok(result.tp1.price > price, "TP1 should be above entry");
    assert.ok(result.tp2.price > result.tp1.price, "TP2 > TP1");
    assert.ok(result.tp3.price > result.tp2.price, "TP3 > TP2");
  });

  it("should calculate SL below entry for BUY", () => {
    const result = calculateTPSL(data, "BUY", price);
    assert.ok(result.sl.price < price, "SL should be below entry");
  });

  it("should include risk/reward ratios", () => {
    const result = calculateTPSL(data, "BUY", price);
    assert.ok(result.riskReward.tp1 >= 1, "TP1 R:R >= 1");
    assert.ok(
      result.riskReward.tp2 > result.riskReward.tp1,
      "TP2 R:R > TP1 R:R"
    );
    assert.ok(
      result.riskReward.tp3 > result.riskReward.tp2,
      "TP3 R:R > TP2 R:R"
    );
  });

  it("should include ATR value", () => {
    const result = calculateTPSL(data, "BUY", price);
    assert.ok(result.atr > 0);
  });

  it("should cap SL at max 6% for STOCK", () => {
    const result = calculateTPSL(data, "BUY", price, "STOCK");
    const slPercent = Math.abs(result.sl.percent);
    assert.ok(slPercent <= 7, `SL ${slPercent}% should be <= ~6%`);
  });
});

// ============================================================
// 3. TP/SL Calculator — SELL (SHORT)
// ============================================================
describe("calculateTPSL — SELL", () => {
  const data = generateOHLC(200, 100);
  const price = data[data.length - 1].close;

  it("should calculate TP1, TP2, TP3 below entry for SELL", () => {
    const result = calculateTPSL(data, "SELL", price, "CRYPTO");
    assert.ok(result.tp1.price < price, "TP1 should be below entry");
    assert.ok(result.tp2.price < result.tp1.price, "TP2 < TP1");
    assert.ok(result.tp3.price < result.tp2.price, "TP3 < TP2");
  });

  it("should calculate SL above entry for SELL", () => {
    const result = calculateTPSL(data, "SELL", price, "CRYPTO");
    assert.ok(result.sl.price > price, "SL should be above entry");
  });
});

// ============================================================
// 4. TP/SL Calculator — Crypto rounding
// ============================================================
describe("calculateTPSL — Crypto Precision", () => {
  it("should use 2dp for prices >= 1000", () => {
    const data = generateOHLC(200, 60000);
    const price = data[data.length - 1].close;
    const result = calculateTPSL(data, "BUY", price, "CRYPTO");
    // Round(x * 100) / 100 → at most 2 decimal places
    const dp = String(result.tp1.price).split(".")[1]?.length || 0;
    assert.ok(dp <= 2, `Expected <= 2dp, got ${dp}`);
  });

  it("should use integer rounding for stock prices >= 1000", () => {
    const data = generateOHLC(200, 8000);
    const price = data[data.length - 1].close;
    const result = calculateTPSL(data, "BUY", price, "STOCK");
    assert.equal(result.tp1.price, Math.round(result.tp1.price));
  });
});
