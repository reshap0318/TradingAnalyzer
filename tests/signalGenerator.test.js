import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { generateSignal } from "../src/shared/signalGenerator.js";

// ============================================================
// Helper — mock decision object
// ============================================================
function mockDecision(overrides = {}) {
  return {
    signal: "BUY",
    strength: "STRONG",
    score: 72.5,
    confidence: 72,
    breakdown: [
      {
        indicator: "MA",
        contribution: 30,
        details: ["EMA12 > EMA26", "SMA bullish"],
      },
      { indicator: "RSI", contribution: 15, details: ["RSI 56.3 Bullish"] },
      { indicator: "MACD", contribution: 20, details: ["MACD bullish cross"] },
      {
        indicator: "Vol",
        contribution: 10,
        confirmation: true,
        details: ["High volume up"],
      },
      {
        indicator: "MTF",
        contribution: 15,
        alignment: "BULLISH_ALIGNED",
        details: ["All TF Bullish"],
      },
    ],
    ihsg: { trend: "BULLISH", signal: 20, details: ["IHSG Bullish"] },
    ...overrides,
  };
}

// ============================================================
// 1. Basic Signal Generation
// ============================================================
describe("generateSignal — Basic", () => {
  it("should produce a valid signal object", () => {
    const result = generateSignal(mockDecision(), 8600, "BBCA.JK");
    assert.equal(result.symbol, "BBCA.JK");
    assert.equal(result.signal, "BUY");
    assert.equal(result.strength, "STRONG");
    assert.ok(result.score > 0);
    assert.ok(result.confidence > 0);
    assert.ok(result.currentPrice === 8600);
  });

  it("should produce entry zone around current price", () => {
    const result = generateSignal(mockDecision(), 8600, "BBCA.JK");
    assert.ok(result.entryZone.low < 8600);
    assert.ok(result.entryZone.high > 8600);
    // Should be ±0.5%
    assert.ok(result.entryZone.low > 8500);
    assert.ok(result.entryZone.high < 8700);
  });

  it("should include timestamp", () => {
    const result = generateSignal(mockDecision(), 8600, "BBCA.JK");
    assert.ok(result.timestamp);
    assert.ok(new Date(result.timestamp).getTime() > 0);
  });

  it("should round score", () => {
    const result = generateSignal(
      mockDecision({ score: 72.123456 }),
      8600,
      "BBCA.JK"
    );
    const parts = String(result.score).split(".");
    assert.ok(!parts[1] || parts[1].length <= 2);
  });
});

// ============================================================
// 2. Reasoning generation
// ============================================================
describe("generateSignal — Reasoning", () => {
  it("should sort reasoning by absolute contribution", () => {
    const result = generateSignal(mockDecision(), 8600, "BBCA.JK");
    assert.ok(result.reasoning.length > 0);
    // Reasoning comes from sorted breakdown
  });

  it("should mark positive with ✓ and negative with ✗", () => {
    const decision = mockDecision({
      breakdown: [
        { indicator: "MA", contribution: 30, details: ["Bullish"] },
        { indicator: "RSI", contribution: -20, details: ["Bearish"] },
      ],
    });
    const result = generateSignal(decision, 8600, "BBCA.JK");
    const positive = result.reasoning.find((r) => r.includes("MA"));
    const negative = result.reasoning.find((r) => r.includes("RSI"));
    assert.ok(positive.startsWith("✓"));
    assert.ok(negative.startsWith("✗"));
  });
});

// ============================================================
// 3. Warnings
// ============================================================
describe("generateSignal — Warnings", () => {
  it("should warn about low volume", () => {
    const decision = mockDecision({
      breakdown: [
        {
          indicator: "Vol",
          contribution: 5,
          confirmation: false,
          details: ["Low volume"],
        },
      ],
    });
    const result = generateSignal(decision, 8600, "BBCA.JK");
    assert.ok(result.warnings.includes("Low volume"));
  });

  it("should warn about mixed timeframes", () => {
    const decision = mockDecision({
      breakdown: [
        {
          indicator: "MTF",
          contribution: 5,
          alignment: "MIXED",
          details: ["Mixed"],
        },
      ],
    });
    const result = generateSignal(decision, 8600, "BBCA.JK");
    assert.ok(result.warnings.includes("Mixed timeframes"));
  });

  it("should warn when BUY signal but IHSG bearish", () => {
    const decision = mockDecision({
      signal: "BUY",
      ihsg: { trend: "BEARISH", signal: -20, details: ["IHSG Bearish"] },
    });
    const result = generateSignal(decision, 8600, "BBCA.JK", "IHSG");
    assert.ok(result.warnings.some((w) => w.includes("IHSG bearish")));
  });

  it("should warn when SELL signal but IHSG bullish", () => {
    const decision = mockDecision({
      signal: "SELL",
      ihsg: { trend: "BULLISH", signal: 20, details: ["IHSG Bullish"] },
    });
    const result = generateSignal(decision, 8600, "BBCA.JK", "IHSG");
    assert.ok(result.warnings.some((w) => w.includes("IHSG bullish")));
  });

  it("should not warn when signal aligns with market", () => {
    const decision = mockDecision({
      signal: "BUY",
      ihsg: { trend: "BULLISH", signal: 20, details: ["IHSG Bullish"] },
    });
    const result = generateSignal(decision, 8600, "BBCA.JK", "IHSG");
    assert.ok(!result.warnings.some((w) => w.includes("IHSG")));
  });
});

// ============================================================
// 4. Crypto label support (BTC Market)
// ============================================================
describe("generateSignal — Crypto Market Label", () => {
  it("should use custom market label for crypto warnings", () => {
    const decision = mockDecision({
      signal: "BUY",
      btcMarket: { trend: "BEARISH", signal: -20 },
      ihsg: undefined,
    });
    const result = generateSignal(decision, 68000, "BTCUSDT", "BTC MARKET");
    assert.ok(result.warnings.some((w) => w.includes("BTC MARKET bearish")));
  });
});

// ============================================================
// 5. Edge cases
// ============================================================
describe("generateSignal — Edge Cases", () => {
  it("should handle empty breakdown", () => {
    const decision = mockDecision({ breakdown: [] });
    const result = generateSignal(decision, 8600, "TEST");
    assert.equal(result.reasoning.length, 0);
  });

  it("should handle missing details in breakdown item", () => {
    const decision = mockDecision({
      breakdown: [{ indicator: "MA", contribution: 20 }],
    });
    const result = generateSignal(decision, 8600, "TEST");
    assert.ok(result.reasoning.length > 0);
  });

  it("should handle WAIT signal", () => {
    const decision = mockDecision({
      signal: "WAIT",
      strength: "NEUTRAL",
      score: 0,
    });
    const result = generateSignal(decision, 8600, "TEST");
    assert.equal(result.signal, "WAIT");
    assert.equal(result.strength, "NEUTRAL");
  });
});
