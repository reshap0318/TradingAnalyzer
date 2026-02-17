import { analyzeATR } from "./indicators/atr.js";
import { findSupportResistance } from "./supportResistance.js";
import { calculateEMA } from "./indicators/movingAverages.js";

export function calculateTPSL(ohlcData, signal, price, assetType = "STOCK") {
  const sr = findSupportResistance(ohlcData);
  const atr = analyzeATR(ohlcData).atr || price * 0.02;
  const closes = ohlcData.map((d) => d.close);
  const ema50Arr = calculateEMA(closes, 50);
  const curEma50 = ema50Arr[ema50Arr.length - 1];

  const result = { tp1: null, tp2: null, tp3: null, sl: null };

  // Crypto uses 8-decimal precision; stocks use integer/1dp/2dp rounding
  const round = (p) => {
    if (assetType === "CRYPTO") {
      // Smart precision: large prices use fewer decimals
      if (p >= 1000) return Math.round(p * 100) / 100;
      if (p >= 1) return Math.round(p * 10000) / 10000;
      return Math.round(p * 100000000) / 100000000;
    }
    return p >= 1000
      ? Math.round(p)
      : p >= 100
      ? Math.round(p * 10) / 10
      : Math.round(p * 100) / 100;
  };

  const maxSlPercent = assetType === "CRYPTO" ? 0.92 : 0.94; // 8% for crypto, 6% for stock

  // Default to LONG (BUY) logic if signal is WAIT
  const isShort = signal === "SELL";

  if (!isShort) {
    // --- STOP LOSS STRATEGY ---
    // 1. Structural SL (Nearest Pivot Low)
    let slDate = "Structure (Pivot)";
    let slPrice = sr.supports.length > 0 ? sr.supports[0].level : null;

    // 2. Dynamic SL (EMA50) - If price is close to EMA50
    if (price > curEma50 && (price - curEma50) / price < 0.05) {
      // If EMA50 is closer than Pivot, maybe use EMA50?
      // User script prefers EMA50 if close.
      slPrice = curEma50;
      slDate = "Dynamic (EMA50)";
    }

    // 3. Volatility SL (2x ATR) - Fallback or if Risk is too small/large
    const atrSl = price - atr * 2;
    if (!slPrice || slPrice < price * maxSlPercent) {
      slPrice = Math.max(slPrice || 0, price * maxSlPercent);
      slDate = `Max Risk (${assetType === "CRYPTO" ? "8" : "6"}%)`;
      if (slPrice < atrSl) {
        // If Max Risk is still wider than ATR, use ATR
        slPrice = atrSl;
        slDate = "Volatility (2x ATR)";
      }
    }

    result.sl = {
      price: round(slPrice),
      percent: ((slPrice - price) / price) * 100,
      reason: slDate,
    };

    // --- TAKE PROFIT STRATEGY (Risk Based) ---
    const riskVal = price - slPrice;

    // Target 1: 1.5R or Nearest Resistance (whichever is higher/safer?)
    // User script: tp1 = max(resistance, price + 1.5R)
    let tp1Price = price + riskVal * 1.5;
    // If resistance is available and > 1R but < 1.5R, maybe use resistance?
    // User script logic: If TP1 (Resistance) < Price + 1R, force TP1 = Price + 1.5R.
    if (sr.resistances[0] && sr.resistances[0].level > price + riskVal) {
      // If resistance is slightly above 1R, use it.
      tp1Price = sr.resistances[0].level;
    }
    // Ensure min 1.5R
    if (tp1Price < price + riskVal * 1.5) tp1Price = price + riskVal * 1.5;

    result.tp1 = {
      price: round(tp1Price),
      percent: ((tp1Price - price) / price) * 100,
      reason: "Risk 1.5x / Res",
    };

    // Target 2: 3R
    // Check if R2 exists and is higher
    let tp2Price = price + riskVal * 3.0;
    if (sr.resistances[1] && sr.resistances[1].level > tp1Price) {
      tp2Price = sr.resistances[1].level;
    }
    if (tp2Price < price + riskVal * 3.0) tp2Price = price + riskVal * 3.0;

    result.tp2 = {
      price: round(tp2Price),
      percent: ((tp2Price - price) / price) * 100,
      reason: "Risk 3.0x / Res",
    };

    // Target 3: 5R (Moonbag)
    let tp3Price = price + riskVal * 5.0;
    result.tp3 = {
      price: round(tp3Price),
      percent: ((tp3Price - price) / price) * 100,
      reason: "Risk 5.0x",
    };
  } else {
    // SHORT LOGIC (Inverted)
    let slDate = "Structure (Pivot)";
    let slPrice = sr.resistances.length > 0 ? sr.resistances[0].level : null;

    // EMA/ATR Logic for Short not prioritized by user script, but let's mirror it
    const atrSl = price + atr * 2;
    const maxSlMultiplierShort = assetType === "CRYPTO" ? 1.08 : 1.06;
    if (!slPrice || slPrice > price * maxSlMultiplierShort) {
      slPrice = Math.min(slPrice || 999999, price * maxSlMultiplierShort);
      slDate = `Max Risk (${assetType === "CRYPTO" ? "8" : "6"}%)`;
      if (slPrice > atrSl) {
        slPrice = atrSl;
        slDate = "Volatility (2x ATR)";
      }
    }
    result.sl = {
      price: round(slPrice),
      percent: ((slPrice - price) / price) * 100,
      reason: slDate,
    };

    const riskVal = slPrice - price;
    let tp1Price = price - riskVal * 1.5;
    if (sr.supports[0] && sr.supports[0].level < price - riskVal)
      tp1Price = sr.supports[0].level;
    if (tp1Price > price - riskVal * 1.5) tp1Price = price - riskVal * 1.5;

    result.tp1 = {
      price: round(tp1Price),
      percent: ((tp1Price - price) / price) * 100,
      reason: "Risk 1.5x",
    };
    result.tp2 = {
      price: round(price - riskVal * 3),
      percent: ((price - riskVal * 3 - price) / price) * 100,
      reason: "Risk 3.0x",
    };
    result.tp3 = {
      price: round(price - riskVal * 5),
      percent: ((price - riskVal * 5 - price) / price) * 100,
      reason: "Risk 5.0x",
    };
  }

  const risk = Math.abs(price - result.sl.price);
  result.riskReward = {
    tp1: Math.round((Math.abs(result.tp1.price - price) / risk) * 100) / 100,
    tp2: Math.round((Math.abs(result.tp2.price - price) / risk) * 100) / 100,
    tp3: Math.round((Math.abs(result.tp3.price - price) / risk) * 100) / 100,
  };
  result.atr = atr;
  result.supports = sr.supports;
  result.resistances = sr.resistances;
  return result;
}
