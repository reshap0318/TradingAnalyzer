import Binance from "binance-api-node";

console.log("--> binanceData.js v3 (RESTORED) LOADED <--");

// Use public client with alternative endpoint for better connectivity
const client = Binance.default({
  httpBase: "https://data-api.binance.vision",
});

/**
 * Fetch Multi-Timeframe Data for Crypto
 * @param {string[]} timeframes - Array of timeframes like ["15m", "1h", "4h", "1d"]
 * @returns {Promise<Object>} - { "15m": [...], "1h": [...], ... }
 */
export async function fetchMultiTimeframe(
  symbol,
  timeframes = ["15m", "1h", "4h", "1d"]
) {
  try {
    const promises = timeframes.map((tf) => {
      // Ensure specific binance intervals, though the map generally aligns closely
      let binanceTf = tf;
      if (tf === "1D") binanceTf = "1d";
      return client.candles({ symbol, interval: binanceTf, limit: 500 });
    });

    const resultsArray = await Promise.all(promises);

    const resultObj = {};
    timeframes.forEach((tf, index) => {
      resultObj[tf] = resultsArray[index].map(formatCandle);
    });

    return resultObj;
  } catch (error) {
    console.error(`Error fetching Binance data for ${symbol}:`, error.message);
    throw error;
  }
}

/**
 * Fetch BTC Dominance or Proxy
 * Since Binance doesn't have a direct "BTC.D" symbol in Spot usually,
 * we check BTCDOMUSDT (Futures) or just return null if fail.
 */
export async function fetchBTCDominance() {
  try {
    // Attempt to fetch BTCDOMUSDT from Futures
    // binance-api-node default client might be spot only unless configured?
    // Actually standard daily candles endpoint works for futures symbols too often,
    // or we might need a specific futures client.
    // Let's try "BTCDOMUSDT".
    const candles = await client.candles({
      symbol: "BTCDOMUSDT",
      interval: "1d",
      limit: 100,
    });

    return candles.map(formatCandle);
  } catch (error) {
    // console.warn(
    //   "Could not fetch BTCDOMUSDT, falling back to BTCUSDT (Market Proxy)."
    // );
    try {
      const btcCandles = await client.candles({
        symbol: "BTCUSDT",
        interval: "1d",
        limit: 100,
      });
      return btcCandles.map(formatCandle);
    } catch (err2) {
      console.error("Failed to fetch market proxy (BTCUSDT):", err2.message);
      return [];
    }
  }
}

/**
 * Get Current Price and stats
 */
export async function getCurrentPrice(symbol) {
  const dailyStats = await client.dailyStats({ symbol });
  // dailyStats is object or array?
  // client.dailyStats({ symbol: 'BTCUSDT' }) returns object.

  return {
    price: parseFloat(dailyStats.lastPrice),
    change: parseFloat(dailyStats.priceChange),
    changePercent: parseFloat(dailyStats.priceChangePercent),
    volume: parseFloat(dailyStats.volume),
    dayHigh: parseFloat(dailyStats.highPrice),
    dayLow: parseFloat(dailyStats.lowPrice),
  };
}

// --- Helper ---

function formatCandle(c) {
  // Binance candle: { openTime, open, high, low, close, volume, ... }
  // App wants: { date, open, high, low, close, volume }
  // Date should be WIB string?
  const d = new Date(c.openTime);
  const wibDate = d.toLocaleString("id-ID", {
    timeZone: "Asia/Jakarta",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });

  return {
    date: wibDate,
    open: parseFloat(c.open),
    high: parseFloat(c.high),
    low: parseFloat(c.low),
    close: parseFloat(c.close),
    volume: parseFloat(c.volume),
  };
}
