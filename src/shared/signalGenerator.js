export function generateSignal(
  decision,
  currentPrice,
  symbol,
  marketLabel = "IHSG"
) {
  const { signal, strength, score, confidence, breakdown } = decision;
  const reasoning = [],
    warnings = [];

  breakdown
    .sort((a, b) => Math.abs(b.contribution) - Math.abs(a.contribution))
    .forEach((item) => {
      reasoning.push(
        `${item.contribution > 0 ? "✓" : "✗"} ${item.indicator}: ${
          item.details?.[0] || ""
        }`
      );
      if (item.indicator === "Vol" && !item.confirmation)
        warnings.push("Low volume");
      if (item.indicator === "MTF" && item.alignment === "MIXED")
        warnings.push("Mixed timeframes");
    });

  // Check market sentiment (IHSG for stocks, BTC Market for crypto)
  const market = decision.ihsg || decision.btcMarket;
  if (market) {
    if (signal === "BUY" && market.trend === "BEARISH")
      warnings.push(`${marketLabel} bearish`);
    else if (signal === "SELL" && market.trend === "BULLISH")
      warnings.push(`${marketLabel} bullish`);
  }

  const buffer = currentPrice * 0.005;
  return {
    symbol,
    signal,
    strength,
    score: Math.round(score * 100) / 100,
    confidence,
    currentPrice,
    entryZone: { low: currentPrice - buffer, high: currentPrice + buffer },
    reasoning,
    warnings,
    timestamp: new Date().toISOString(),
  };
}
