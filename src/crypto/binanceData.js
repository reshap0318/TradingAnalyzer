import Binance from "binance-api-node";

console.log("--> binanceData.js v3 (RESTORED) LOADED <--");

// Use public client with alternative endpoint for better connectivity
const client = Binance.default({
  httpBase: "https://data-api.binance.vision",
});

/**
 * Fetch Multi-Timeframe Data for Crypto
 * @param {string} symbol - e.g. "BTCUSDT"
 * @returns {Promise<Object>} - { "1D": [...], "4h": [...], "1h": [...] }
 */
export async function fetchMultiTimeframe(symbol) {
  try {
    // Fetch 1h, 4h, 1d, 15m candles
    // Limit 500 candles is standard
    const [candles1d, candles4h, candles1h, candles15m] = await Promise.all([
      client.candles({ symbol, interval: "1d", limit: 500 }),
      client.candles({ symbol, interval: "4h", limit: 500 }),
      client.candles({ symbol, interval: "1h", limit: 500 }),
      client.candles({ symbol, interval: "15m", limit: 500 }),
    ]);

    return {
      "1D": candles1d.map(formatCandle),
      "4h": candles4h.map(formatCandle),
      "1h": candles1h.map(formatCandle),
      "15m": candles15m.map(formatCandle),
    };
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
    console.warn(
      "Could not fetch BTCDOMUSDT, falling back to BTCUSDT (Market Proxy)."
    );
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
