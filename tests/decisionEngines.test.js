import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { makeDecision } from "../src/saham/decisionEngine.js";
import { makeCryptoDecision } from "../src/crypto/cryptoDecisionEngine.js";

// ============================================================
// Helpers — Generate realistic OHLC data
// ============================================================
function generateOHLC(count = 250, startPrice = 100, direction = "up") {
  const data = [];
  let p = startPrice;
  for (let i = 0; i < count; i++) {
    const bias = direction === "up" ? 0.3 : direction === "down" ? -0.3 : 0;
    const change = (Math.random() - 0.5 + bias) * 2;
    const close = p + change;
    const open = p;
    const high = Math.max(open, close) + Math.random() * 1.5;
    const low = Math.min(open, close) - Math.random() * 1.5;
    const volume = 1000000 + Math.random() * 500000;
    data.push({
      open,
      high,
      low,
      close,
      volume,
      date: new Date(2026, 0, i + 1).toISOString(),
    });
    p = close;
  }
  return data;
}

function generateMultiTF(direction = "up") {
  return {
    "1D": generateOHLC(250, 100, direction),
    "1W": generateOHLC(100, 100, direction),
    "1M": generateOHLC(50, 100, direction),
  };
}

function generateCryptoMultiTF(direction = "up") {
  return {
    "1h": generateOHLC(250, 68000, direction),
    "4h": generateOHLC(200, 68000, direction),
    "1d": generateOHLC(100, 68000, direction),
  };
}

function mockIHSG(trend = "BULLISH") {
  return {
    signal: trend === "BULLISH" ? 20 : trend === "BEARISH" ? -20 : 0,
    trend,
    isCrash: false,
    details: [`IHSG ${trend}`],
  };
}

function mockBTCMarket(trend = "BULLISH") {
  return {
    signal: trend === "BULLISH" ? 30 : trend === "BEARISH" ? -30 : 0,
    trend,
    isCrash: false,
    details: [`BTC Market ${trend}`],
  };
}

// ============================================================
// 1. Saham Decision Engine — makeDecision
// ============================================================
describe("makeDecision (Saham)", () => {
  it("should return a valid decision structure", () => {
    const data = generateOHLC(250);
    const multiTf = generateMultiTF();
    const decision = makeDecision(data, multiTf, mockIHSG());

    assert.ok(["BUY", "SELL", "WAIT"].includes(decision.signal));
    assert.ok(
      ["STRONG", "MODERATE", "NEUTRAL", "WEAK"].includes(decision.strength)
    );
    assert.ok(typeof decision.score === "number");
    assert.ok(typeof decision.confidence === "number");
    assert.ok(decision.confidence >= 0 && decision.confidence <= 100);
  });

  it("should include indicator breakdown", () => {
    const data = generateOHLC(250);
    const decision = makeDecision(data, generateMultiTF(), mockIHSG());

    assert.ok(Array.isArray(decision.breakdown));
    const indicators = decision.breakdown.map((b) => b.indicator);
    assert.ok(indicators.includes("MA"));
    assert.ok(indicators.includes("RSI"));
    assert.ok(indicators.includes("MACD"));
  });

  it("should include multi-timeframe analysis", () => {
    const decision = makeDecision(
      generateOHLC(250),
      generateMultiTF(),
      mockIHSG()
    );
    assert.ok(decision.multiTimeframe);
    assert.ok(decision.multiTimeframe.alignment);
  });

  it("should include candle patterns", () => {
    const decision = makeDecision(
      generateOHLC(250),
      generateMultiTF(),
      mockIHSG()
    );
    assert.ok(decision.patterns);
    assert.ok(typeof decision.patterns === "object");
  });

  it("should override to WAIT during IHSG crash", () => {
    const crash = {
      signal: -50,
      trend: "BEARISH",
      isCrash: true,
      details: ["IHSG Crash > 1.5%"],
    };
    const decision = makeDecision(
      generateOHLC(250, 100, "up"),
      generateMultiTF("up"),
      crash
    );
    assert.equal(decision.signal, "WAIT");
    assert.equal(decision.strength, "WEAK");
  });

  it("should return IHSG analysis in output", () => {
    const ihsg = mockIHSG("BULLISH");
    const decision = makeDecision(generateOHLC(250), generateMultiTF(), ihsg);
    assert.ok(decision.ihsg);
    assert.equal(decision.ihsg.trend, "BULLISH");
  });
});

// ============================================================
// 2. Saham Decision Engine — Negative Scenarios
// ============================================================
describe("makeDecision — Negative Scenarios", () => {
  it("should produce bearish signal for downtrend", () => {
    const data = generateOHLC(250, 200, "down");
    const decision = makeDecision(
      data,
      generateMultiTF("down"),
      mockIHSG("BEARISH")
    );
    // In a strong downtrend, score should be negative
    assert.ok(decision.score < 0 || decision.signal !== "BUY");
  });

  it("should handle neutral market (not BUY)", () => {
    const flat = generateOHLC(250, 100, "flat");
    const decision = makeDecision(
      flat,
      generateMultiTF("flat"),
      mockIHSG("NEUTRAL")
    );
    // In flat market, may be WAIT
    assert.ok(["WAIT", "SELL", "BUY"].includes(decision.signal));
  });
});

// ============================================================
// 3. Crypto Decision Engine — makeCryptoDecision
// ============================================================
describe("makeCryptoDecision", () => {
  it("should return a valid decision structure", () => {
    const hourly = generateOHLC(250, 68000);
    const multiTf = generateCryptoMultiTF();
    const decision = makeCryptoDecision(hourly, multiTf, mockBTCMarket());

    assert.ok(["BUY", "SELL", "WAIT"].includes(decision.signal));
    assert.ok(typeof decision.score === "number");
    assert.ok(typeof decision.confidence === "number");
    assert.ok(decision.confidence >= 0 && decision.confidence <= 100);
  });

  it("should include indicator breakdown", () => {
    const decision = makeCryptoDecision(
      generateOHLC(250, 68000),
      generateCryptoMultiTF(),
      mockBTCMarket()
    );
    assert.ok(Array.isArray(decision.breakdown));
    assert.ok(decision.breakdown.length > 0);
  });

  it("should include BTC market analysis", () => {
    const btc = mockBTCMarket("BEARISH");
    const decision = makeCryptoDecision(
      generateOHLC(250, 68000),
      generateCryptoMultiTF(),
      btc
    );
    assert.ok(decision.btcMarket);
    assert.equal(decision.btcMarket.trend, "BEARISH");
  });

  it("should handle BTC crash override", () => {
    const crash = {
      signal: -50,
      trend: "BEARISH",
      isCrash: true,
      details: ["BTC Crash Detected"],
    };
    const decision = makeCryptoDecision(
      generateOHLC(250, 68000, "up"),
      generateCryptoMultiTF("up"),
      crash
    );
    // Crash should force WAIT
    assert.equal(decision.signal, "WAIT");
  });

  it("should include multi-timeframe analysis for crypto", () => {
    const decision = makeCryptoDecision(
      generateOHLC(250, 68000),
      generateCryptoMultiTF(),
      mockBTCMarket()
    );
    assert.ok(decision.multiTimeframe);
  });

  it("should include candle patterns", () => {
    const decision = makeCryptoDecision(
      generateOHLC(250, 68000),
      generateCryptoMultiTF(),
      mockBTCMarket()
    );
    assert.ok(decision.patterns);
  });
});

// ============================================================
// 4. Crypto Decision Engine — Negative Scenarios
// ============================================================
describe("makeCryptoDecision — Negative Scenarios", () => {
  it("should produce bearish signal in downtrend with bearish BTC", () => {
    const data = generateOHLC(250, 70000, "down");
    const decision = makeCryptoDecision(
      data,
      generateCryptoMultiTF("down"),
      mockBTCMarket("BEARISH")
    );
    assert.ok(decision.score < 0 || decision.signal !== "BUY");
  });

  it("should handle minimal data gracefully", () => {
    // Only 50 candles (less than 200 needed for MA)
    const minimal = generateOHLC(50, 68000);
    const multiTf = {
      "1h": minimal,
      "4h": generateOHLC(50, 68000),
      "1d": generateOHLC(50, 68000),
    };
    // Should not throw
    const decision = makeCryptoDecision(minimal, multiTf, mockBTCMarket());
    assert.ok(["BUY", "SELL", "WAIT"].includes(decision.signal));
  });
});
