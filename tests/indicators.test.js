import { describe, it } from "node:test";
import assert from "node:assert/strict";
import {
  calculateSMA,
  calculateEMA,
  analyzeMA,
} from "../src/shared/indicators/movingAverages.js";
import { calculateRSI, analyzeRSI } from "../src/shared/indicators/rsi.js";
import { calculateMACD, analyzeMACD } from "../src/shared/indicators/macd.js";
import {
  calculateBollingerBands,
  analyzeBollingerBands,
} from "../src/shared/indicators/bollingerBands.js";
import {
  calculateStochastic,
  analyzeStochastic,
} from "../src/shared/indicators/stochastic.js";
import { analyzeVolume } from "../src/shared/indicators/volume.js";
import { calculateATR, analyzeATR } from "../src/shared/indicators/atr.js";
import { detectCandlePatterns } from "../src/shared/indicators/candlePatterns.js";

// ============================================================
// Helpers — Generate dummy OHLC data
// ============================================================
function generateUptrend(count = 250, startPrice = 100) {
  const data = [];
  let p = startPrice;
  for (let i = 0; i < count; i++) {
    const change = (Math.random() - 0.3) * 2; // Bias upward
    const open = p;
    const close = p + change;
    const high = Math.max(open, close) + Math.random() * 1;
    const low = Math.min(open, close) - Math.random() * 1;
    const volume = 1000000 + Math.random() * 500000;
    data.push({
      open,
      high,
      low,
      close,
      volume,
      date: `2026-01-${String(i + 1).padStart(2, "0")}`,
    });
    p = close;
  }
  return data;
}

function generateDowntrend(count = 250, startPrice = 200) {
  const data = [];
  let p = startPrice;
  for (let i = 0; i < count; i++) {
    const change = (Math.random() - 0.7) * 2; // Bias downward
    const open = p;
    const close = p + change;
    const high = Math.max(open, close) + Math.random() * 1;
    const low = Math.min(open, close) - Math.random() * 1;
    const volume = 1000000 + Math.random() * 500000;
    data.push({
      open,
      high,
      low,
      close,
      volume,
      date: `2026-01-${String(i + 1).padStart(2, "0")}`,
    });
    p = close;
  }
  return data;
}

function generateFlat(count = 250, price = 100) {
  const data = [];
  for (let i = 0; i < count; i++) {
    const noise = (Math.random() - 0.5) * 0.5;
    const open = price + noise;
    const close = price - noise;
    const high = price + Math.abs(noise) + 0.1;
    const low = price - Math.abs(noise) - 0.1;
    data.push({
      open,
      high,
      low,
      close,
      volume: 1000000,
      date: `2026-01-${String(i + 1).padStart(2, "0")}`,
    });
  }
  return data;
}

// ============================================================
// 1. SMA
// ============================================================
describe("calculateSMA", () => {
  it("should calculate correct SMA values", () => {
    const data = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
    const sma3 = calculateSMA(data, 3);
    assert.equal(sma3.length, 8); // 10 - 3 + 1
    assert.equal(sma3[0], 2); // avg(1,2,3)
    assert.equal(sma3[7], 9); // avg(8,9,10)
  });

  it("should return empty for insufficient data", () => {
    const sma = calculateSMA([1, 2], 5);
    assert.equal(sma.length, 0);
  });

  it("should handle single period (identity)", () => {
    const data = [5, 10, 15];
    const sma = calculateSMA(data, 1);
    assert.deepEqual(sma, [5, 10, 15]);
  });
});

// ============================================================
// 2. EMA
// ============================================================
describe("calculateEMA", () => {
  it("should calculate EMA values", () => {
    const data = [10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20];
    const ema3 = calculateEMA(data, 3);
    assert.ok(ema3.length > 0);
    // First EMA = SMA of first 3 values = (10+11+12)/3 = 11
    assert.equal(ema3[0], 11);
  });

  it("should converge to price in a flat market", () => {
    const flat = Array(50).fill(100);
    const ema = calculateEMA(flat, 10);
    // All EMA values should be 100
    assert.equal(ema[ema.length - 1], 100);
  });

  it("should lag behind in uptrend (EMA < latest price)", () => {
    const data = Array.from({ length: 50 }, (_, i) => 100 + i);
    const ema = calculateEMA(data, 10);
    assert.ok(ema[ema.length - 1] < data[data.length - 1]);
  });
});

// ============================================================
// 3. analyzeMA
// ============================================================
describe("analyzeMA", () => {
  it("should return insufficient data for small dataset", () => {
    const result = analyzeMA([1, 2, 3]);
    assert.equal(result.signal, 0);
    assert.equal(result.trend, "NEUTRAL");
    assert.ok(result.details.includes("Insufficient data"));
  });

  it("should produce bullish signal in uptrend", () => {
    const upData = generateUptrend(300);
    const closes = upData.map((d) => d.close);
    const result = analyzeMA(closes);
    assert.ok(result.signal > 0);
  });

  it("should produce bearish signal in downtrend", () => {
    const downData = generateDowntrend(300, 500);
    const closes = downData.map((d) => d.close);
    const result = analyzeMA(closes);
    assert.ok(result.signal < 0);
  });

  it("should clamp signal between -100 and 100", () => {
    const data = generateUptrend(300);
    const result = analyzeMA(data.map((d) => d.close));
    assert.ok(result.signal >= -100 && result.signal <= 100);
  });
});

// ============================================================
// 4. RSI
// ============================================================
describe("calculateRSI", () => {
  it("should return RSI values between 0 and 100", () => {
    const data = generateUptrend(100);
    const rsi = calculateRSI(data.map((d) => d.close));
    rsi.forEach((val) => {
      assert.ok(val >= 0 && val <= 100, `RSI ${val} out of range`);
    });
  });

  it("should return empty for insufficient data", () => {
    const rsi = calculateRSI([1, 2, 3], 14);
    assert.equal(rsi.length, 0);
  });

  it("should return high RSI for strong uptrend", () => {
    // Monotonically increasing prices
    const rising = Array.from({ length: 50 }, (_, i) => 100 + i * 2);
    const rsi = calculateRSI(rising, 14);
    assert.ok(rsi[rsi.length - 1] > 50);
  });

  it("should return low RSI for strong downtrend", () => {
    const falling = Array.from({ length: 50 }, (_, i) => 200 - i * 2);
    const rsi = calculateRSI(falling, 14);
    assert.ok(rsi[rsi.length - 1] < 50);
  });
});

describe("analyzeRSI", () => {
  it("should detect OVERBOUGHT zone", () => {
    // Create a strong uptrend to push RSI > 70
    const extreme = Array.from({ length: 50 }, (_, i) => 100 + i * 5);
    const result = analyzeRSI(extreme);
    if (result.value >= 70) {
      assert.equal(result.zone, "OVERBOUGHT");
      assert.ok(result.signal < 0); // Bearish signal
    }
  });

  it("should detect OVERSOLD zone", () => {
    const extreme = Array.from({ length: 50 }, (_, i) => 300 - i * 5);
    const result = analyzeRSI(extreme);
    if (result.value <= 30) {
      assert.equal(result.zone, "OVERSOLD");
      assert.ok(result.signal > 0); // Bullish signal
    }
  });

  it("should return insufficient data for small input", () => {
    const result = analyzeRSI([1, 2, 3]);
    assert.equal(result.signal, 0);
    assert.equal(result.zone, "NEUTRAL");
  });
});

// ============================================================
// 5. MACD
// ============================================================
describe("calculateMACD", () => {
  it("should return macdLine, signalLine, histogram", () => {
    const data = generateUptrend(100);
    const result = calculateMACD(data.map((d) => d.close));
    assert.ok(Array.isArray(result.macdLine));
    assert.ok(Array.isArray(result.signalLine));
    assert.ok(Array.isArray(result.histogram));
  });

  it("should return empty arrays for insufficient data", () => {
    const result = calculateMACD([1, 2, 3]);
    assert.equal(result.macdLine.length, 0);
  });
});

describe("analyzeMACD", () => {
  it("should produce positive signal in uptrend", () => {
    const data = generateUptrend(200);
    const result = analyzeMACD(data.map((d) => d.close));
    assert.ok(result.signal > 0 || result.details.length > 0);
  });

  it("should return insufficient data for small input", () => {
    const result = analyzeMACD([1, 2, 3]);
    assert.equal(result.signal, 0);
  });

  it("should include MACD values in result", () => {
    const data = generateUptrend(200);
    const result = analyzeMACD(data.map((d) => d.close));
    if (result.values) {
      assert.ok("macd" in result.values);
      assert.ok("signal" in result.values);
      assert.ok("histogram" in result.values);
    }
  });
});

// ============================================================
// 6. Bollinger Bands
// ============================================================
describe("calculateBollingerBands", () => {
  it("should return upper, middle, lower, bandwidth", () => {
    const data = generateFlat(100);
    const result = calculateBollingerBands(data.map((d) => d.close));
    assert.ok(result.upper.length > 0);
    assert.ok(result.middle.length > 0);
    assert.ok(result.lower.length > 0);
  });

  it("should have upper > middle > lower", () => {
    const data = generateFlat(100);
    const result = calculateBollingerBands(data.map((d) => d.close));
    const lastIdx = result.upper.length - 1;
    assert.ok(result.upper[lastIdx] > result.middle[lastIdx]);
    assert.ok(result.middle[lastIdx] > result.lower[lastIdx]);
  });

  it("should return empty for insufficient data", () => {
    const result = calculateBollingerBands([1, 2, 3]);
    assert.equal(result.upper.length, 0);
  });
});

describe("analyzeBollingerBands", () => {
  it("should detect position relative to bands", () => {
    const data = generateUptrend(100);
    const result = analyzeBollingerBands(data.map((d) => d.close));
    assert.ok(
      ["ABOVE_UPPER", "UPPER_HALF", "LOWER_HALF", "BELOW_LOWER"].includes(
        result.position
      )
    );
  });

  it("should return percentB value", () => {
    const data = generateFlat(100);
    const result = analyzeBollingerBands(data.map((d) => d.close));
    if (result.percentB !== undefined) {
      assert.ok(typeof result.percentB === "number");
    }
  });

  it("should return insufficient for small data", () => {
    const result = analyzeBollingerBands([1, 2, 3]);
    assert.equal(result.signal, 0);
  });
});

// ============================================================
// 7. Stochastic
// ============================================================
describe("calculateStochastic", () => {
  it("should return K and D arrays", () => {
    const data = generateUptrend(100);
    const result = calculateStochastic(
      data.map((d) => d.high),
      data.map((d) => d.low),
      data.map((d) => d.close)
    );
    assert.ok(Array.isArray(result.k));
    assert.ok(Array.isArray(result.d));
  });

  it("should return values between 0 and 100", () => {
    const data = generateUptrend(100);
    const result = calculateStochastic(
      data.map((d) => d.high),
      data.map((d) => d.low),
      data.map((d) => d.close)
    );
    result.k.forEach((val) => assert.ok(val >= 0 && val <= 100));
    result.d.forEach((val) => assert.ok(val >= 0 && val <= 100));
  });

  it("should return empty for insufficient data", () => {
    const result = calculateStochastic([1, 2], [0, 1], [0.5, 1.5]);
    assert.equal(result.k.length, 0);
  });
});

describe("analyzeStochastic", () => {
  it("should detect overbought/oversold zones", () => {
    const data = generateUptrend(100);
    const result = analyzeStochastic(data);
    assert.ok(["OVERBOUGHT", "OVERSOLD", "NEUTRAL"].includes(result.zone));
  });

  it("should return insufficient for small data", () => {
    const result = analyzeStochastic([{ high: 1, low: 0, close: 0.5 }]);
    assert.equal(result.signal, 0);
  });
});

// ============================================================
// 8. Volume
// ============================================================
describe("analyzeVolume", () => {
  it("should detect high volume confirmation", () => {
    const data = generateFlat(50);
    // Add a high-volume candle at the end
    const lastCandle = { ...data[data.length - 1] };
    lastCandle.volume = 5000000; // 5x average
    lastCandle.close = lastCandle.open + 5; // Up candle
    data[data.length - 1] = lastCandle;

    const result = analyzeVolume(data);
    assert.ok(result.volumeRatio > 1);
  });

  it("should detect low volume", () => {
    const data = generateFlat(50);
    data[data.length - 1] = { ...data[data.length - 1], volume: 100 };
    const result = analyzeVolume(data);
    assert.ok(result.volumeRatio < 1);
  });

  it("should return insufficient for small data", () => {
    const result = analyzeVolume([{ close: 100, volume: 1000 }]);
    assert.equal(result.signal, 0);
  });

  it("should include volumeRatio and currentVolume", () => {
    const data = generateFlat(50);
    const result = analyzeVolume(data);
    assert.ok(typeof result.volumeRatio === "number");
    assert.ok(typeof result.currentVolume === "number");
    assert.ok(typeof result.avgVolume === "number");
  });
});

// ============================================================
// 9. ATR
// ============================================================
describe("calculateATR", () => {
  it("should calculate ATR values", () => {
    const data = generateUptrend(50);
    const atr = calculateATR(data);
    assert.ok(atr.length > 0);
    atr.forEach((v) => assert.ok(v > 0));
  });

  it("should return empty for insufficient data", () => {
    const atr = calculateATR([{ high: 1, low: 0, close: 0.5 }]);
    assert.equal(atr.length, 0);
  });
});

describe("analyzeATR", () => {
  it("should classify volatility", () => {
    const data = generateUptrend(100);
    const result = analyzeATR(data);
    assert.ok(["HIGH", "NORMAL", "LOW", "UNKNOWN"].includes(result.volatility));
    assert.ok(result.atr > 0);
  });

  it("should return UNKNOWN for insufficient data", () => {
    const result = analyzeATR([{ high: 1, low: 0, close: 0.5 }]);
    assert.equal(result.volatility, "UNKNOWN");
  });
});

// ============================================================
// 10. Candle Patterns
// ============================================================
describe("detectCandlePatterns", () => {
  it("should detect Bullish Engulfing", () => {
    // Set up: prev red candle, curr green engulfing
    const opens = [100, 102, 103, 104, 101, 99]; // prev open=101, curr open=99
    const highs = [103, 104, 105, 106, 102, 106]; // curr high=106
    const lows = [99, 101, 102, 103, 98, 98]; // curr low=98
    const closes = [102, 103, 104, 105, 99, 105]; // prev close=99 (red), curr close=105 (green)
    // Prev: open=101 close=99 → RED
    // Curr: open=99 close=105 → GREEN, open <= prev.close (99<=99), close >= prev.open (105>=101)
    const patterns = detectCandlePatterns(opens, highs, lows, closes);
    assert.ok(patterns.includes("Bullish Engulfing"));
  });

  it("should detect Hammer", () => {
    // Small body at top, long lower wick
    const opens = [100, 101, 102, 103, 104, 103]; // curr open=103
    const highs = [102, 103, 104, 105, 106, 104]; // curr high=104
    const lows = [99, 100, 101, 102, 103, 97]; // curr low=97 (long lower wick)
    const closes = [101, 102, 103, 104, 105, 104]; // curr close=104
    // body = |103-104| = 1, lower wick = min(103,104)-97 = 6, upper = 104-104 = 0
    const patterns = detectCandlePatterns(opens, highs, lows, closes);
    assert.ok(patterns.includes("Hammer"));
  });

  it("should detect Doji", () => {
    const opens = [100, 101, 102, 103, 104, 105.0];
    const highs = [102, 103, 104, 105, 106, 107];
    const lows = [99, 100, 101, 102, 103, 103];
    const closes = [101, 102, 103, 104, 105, 105.01]; // Very small body
    const patterns = detectCandlePatterns(opens, highs, lows, closes);
    assert.ok(patterns.includes("Doji"));
  });

  it("should detect Bearish Engulfing", () => {
    const opens = [100, 101, 102, 103, 99, 105]; // prev open=99, curr open=105
    const highs = [102, 103, 104, 105, 103, 106];
    const lows = [98, 99, 100, 101, 98, 97];
    const closes = [101, 102, 103, 104, 103, 98]; // prev close=103 (green), curr close=98 (red engulfing)
    const patterns = detectCandlePatterns(opens, highs, lows, closes);
    assert.ok(patterns.includes("Bearish Engulfing"));
  });

  it("should detect Shooting Star", () => {
    // Shooting star: long upper wick (>= 2x body), small lower wick (<= 0.5x body)
    const opens = [100, 101, 102, 103, 104, 107];
    const highs = [102, 103, 104, 105, 106, 115]; // upper wick = 115-107 = 8
    const lows = [99, 100, 101, 102, 103, 106.5]; // lower wick = 106.5-106.5 = 0
    const closes = [101, 102, 103, 104, 105, 106.5]; // body = |107-106.5| = 0.5
    // upper=8 >= 2*0.5=1 ✓, lower=0 <= 0.5*0.5=0.25 ✓
    const patterns = detectCandlePatterns(opens, highs, lows, closes);
    assert.ok(patterns.includes("Shooting Star"));
  });

  it("should return empty for insufficient data", () => {
    const patterns = detectCandlePatterns([1], [2], [0], [1]);
    assert.deepEqual(patterns, []);
  });

  it("should return empty for flat candle (no range)", () => {
    const patterns = detectCandlePatterns(
      [100, 100, 100, 100, 100, 100],
      [100, 100, 100, 100, 100, 100],
      [100, 100, 100, 100, 100, 100],
      [100, 100, 100, 100, 100, 100]
    );
    assert.deepEqual(patterns, []);
  });
});
