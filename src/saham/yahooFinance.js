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

export async function fetchMultiTimeframe(
  symbol,
  timeframes = ["15m", "1h", "4h", "1d"]
) {
  const resultObj = {};

  try {
    // Determine required Yahoo intervals
    const fetchPromises = [];
    const fetchMaps = [];

    // Map requested to Yahoo supported
    const yfMap = {
      "1m": "1m",
      "5m": "5m",
      "15m": "15m",
      "30m": "30m",
      "1h": "1h",
      "1d": "1d",
      "1w": "1wk",
      "1M": "1mo",
    };

    // Period mapping
    const getDaysForTf = (tf) => {
      if (tf.endsWith("m")) return 7; // Intraday limits
      if (tf.endsWith("h")) return 30;
      return 400; // Daily/Weekly/Monthly
    };

    // If '4h' or '2h' is requested, we need '1h' to aggregate
    let needsHourly = timeframes.some((tf) =>
      ["2h", "4h", "6h", "8h", "12h"].includes(tf)
    );
    let tfsToFetch = [...timeframes];
    if (needsHourly && !tfsToFetch.includes("1h")) tfsToFetch.push("1h");

    for (const tf of tfsToFetch) {
      if (yfMap[tf]) {
        fetchPromises.push(
          yahooFinance.chart(symbol, {
            period1: getDateStr(getDaysForTf(tf)),
            interval: yfMap[tf],
          })
        );
        fetchMaps.push(tf);
      }
    }

    const fetchedData = await Promise.all(fetchPromises);

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
    };

    const parsedTfs = {};
    fetchMaps.forEach((tf, idx) => {
      parsedTfs[tf] = parseQuotes(fetchedData[idx]);
    });

    // Handle standard requested outputs
    for (const tf of timeframes) {
      if (parsedTfs[tf]) {
        resultObj[tf] = parsedTfs[tf];
      } else if (
        ["2h", "4h", "6h", "8h", "12h"].includes(tf) &&
        parsedTfs["1h"]
      ) {
        // Fallback or aggregate
        if (tf === "4h") {
          resultObj[tf] = aggregate4hFromHourly(parsedTfs["1h"]);
        } else {
          // For other unsupported ones, just use 1h for now or empty
          resultObj[tf] = [];
        }
      } else {
        resultObj[tf] = []; // Unsupported
      }
    }

    return resultObj;
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
