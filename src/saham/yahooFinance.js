/**
 * Yahoo Finance Data Fetcher (ES Module)
 */
import YahooFinance from "yahoo-finance2";

const yahooFinance = new YahooFinance();

const getDateStr = (daysAgo) => {
  const d = new Date();
  d.setDate(d.getDate() - daysAgo);
  return d.toISOString().split("T")[0];
};

export async function fetchMultiTimeframe(symbol) {
  // Define start dates for each timeframe
  const startDaily = getDateStr(400); // ~1 year for daily (EMA200)
  const start4h = getDateStr(90); // 3 months for 4h
  const startHourly = getDateStr(30); // 1 month for hourly
  const start15m = getDateStr(7); // 7 days for 15m (Yahoo max ~60 days for 15m)

  try {
    // Fetch all timeframes from Yahoo Finance
    const [daily, hourly, min15] = await Promise.all([
      yahooFinance.chart(symbol, { period1: startDaily, interval: "1d" }),
      yahooFinance.chart(symbol, { period1: startHourly, interval: "1h" }),
      yahooFinance.chart(symbol, { period1: start15m, interval: "15m" }),
    ]);

    const parseQuotes = (result) => {
      if (!result || !result.quotes || result.quotes.length === 0) return [];
      return result.quotes
        .map((q) => {
          const dateObj = new Date(q.date);
          const wibDate = dateObj.toLocaleString("id-ID", {
            timeZone: "Asia/Jakarta",
            year: "numeric",
            month: "2-digit",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
            hour12: false,
          }); // e.g. "14/02/2026, 14:00:00" (format varies by locale, let's trust it or enforce)

          return {
            date: wibDate, // User wants this to be Indonesia time
            open: q.open,
            high: q.high,
            low: q.low,
            close: q.close,
            volume: q.volume,
          };
        })
        .filter((q) => q.open && q.high && q.low && q.close);
    };

    // Convert 1h data to 4h by aggregating every 4 candles
    const hourlyData = parseQuotes(hourly);
    const data4h = aggregate4hFromHourly(hourlyData);

    return {
      "1D": parseQuotes(daily),
      "4h": data4h,
      "1h": hourlyData,
      "15m": parseQuotes(min15),
    };
  } catch (error) {
    console.error(
      `Error fetching multi-timeframe for ${symbol}:`,
      error.message
    );
    return { "1D": [], "4h": [], "1h": [], "15m": [] };
  }
}

/**
 * Aggregate hourly data into 4h candles
 */
function aggregate4hFromHourly(hourlyData) {
  if (!hourlyData || hourlyData.length === 0) return [];

  const aggregated = [];
  let currentCandle = null;

  hourlyData.forEach((c) => {
    // Parse WIB String format: "13/02/2026, 16.00.00"
    // Split by ", " -> ["13/02/2026", "16.00.00"]
    const parts = c.date.split(", ");
    if (parts.length < 2) return; // Skip invalid formats

    const datePart = parts[0]; // "13/02/2026"
    const timePart = parts[1]; // "16.00.00"

    // Extract Hour
    const hour = parseInt(timePart.split(".")[0], 10);

    // Define Sessions:
    // Session 1: 09:00 - 12:59 -> Key suffix "S1", Time "09.00.00"
    // Session 2: 13:00 - 16:59 -> Key suffix "S2", Time "13.00.00"
    let sessionKey, sessionTime;

    if (hour < 13) {
      sessionKey = "S1";
      sessionTime = "09.00.00";
    } else {
      sessionKey = "S2";
      sessionTime = "13.00.00";
    }

    const key = `${datePart}-${sessionKey}`;

    if (!currentCandle || currentCandle.key !== key) {
      if (currentCandle) aggregated.push(currentCandle.data);
      currentCandle = {
        key,
        data: {
          date: `${datePart}, ${sessionTime}`, // Force start time of session
          open: c.open,
          high: c.high,
          low: c.low,
          close: c.close,
          volume: c.volume || 0,
        },
      };
    } else {
      // Update
      currentCandle.data.high = Math.max(currentCandle.data.high, c.high);
      currentCandle.data.low = Math.min(currentCandle.data.low, c.low);
      currentCandle.data.close = c.close; // Latest close
      currentCandle.data.volume += c.volume || 0;
    }
  });

  // Push last candle
  if (currentCandle) aggregated.push(currentCandle.data);

  return aggregated;
}

export async function fetchIHSG() {
  const startDaily = getDateStr(120);
  try {
    const result = await yahooFinance.chart("^JKSE", {
      period1: startDaily,
      interval: "1d",
    });
    if (!result || !result.quotes) return [];
    return result.quotes
      .map((q) => {
        const dateObj = new Date(q.date);
        const wibDate = dateObj.toLocaleString("id-ID", {
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
          open: q.open,
          high: q.high,
          low: q.low,
          close: q.close,
          volume: q.volume,
        };
      })
      .filter((q) => q.open && q.high && q.low && q.close);
  } catch (error) {
    console.error("Error fetching IHSG:", error.message);
    return [];
  }
}

export async function getCurrentPrice(symbol) {
  try {
    const quote = await yahooFinance.quote(symbol);
    return {
      price: quote.regularMarketPrice,
      change: quote.regularMarketChange,
      changePercent: quote.regularMarketChangePercent,
      volume: quote.regularMarketVolume,
      dayHigh: quote.regularMarketDayHigh,
      dayLow: quote.regularMarketDayLow,
    };
  } catch (error) {
    console.error(`Error fetching quote for ${symbol}:`, error.message);
    throw error;
  }
}

export async function fetchOHLCV(symbol, interval = "1d", daysAgo = 180) {
  const period1 = getDateStr(daysAgo);
  try {
    const result = await yahooFinance.chart(symbol, { period1, interval });
    if (!result || !result.quotes) return [];
    return result.quotes
      .map((q) => ({
        date: q.date,
        open: q.open,
        high: q.high,
        low: q.low,
        close: q.close,
        volume: q.volume,
      }))
      .filter((q) => q.open && q.high && q.low && q.close);
  } catch (error) {
    console.error(`Error fetching ${symbol}:`, error.message);
    throw error;
  }
}
