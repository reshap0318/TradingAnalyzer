import Binance from "binance-api-node";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
import dotenv from "dotenv";
import config from "../config.js";

dotenv.config();

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ORDERS_FILE = path.join(__dirname, "../../data/binance_orders.json");

let client;

// ─── Local JSON DB Management ───
export function loadBinanceOrders() {
  if (!fs.existsSync(ORDERS_FILE)) {
    saveBinanceOrders({ testnet: [], prod: [] });
  }
  return JSON.parse(fs.readFileSync(ORDERS_FILE, "utf-8"));
}

export function saveBinanceOrders(data) {
  const dir = path.dirname(ORDERS_FILE);
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(ORDERS_FILE, JSON.stringify(data, null, 2), "utf-8");
}

export function saveBinanceOrder(orderData) {
  const db = loadBinanceOrders();
  const network = config.AUTO_TRADING.USE_TESTNET ? "testnet" : "prod";
  if (!db[network]) db[network] = [];
  db[network].push(orderData);
  saveBinanceOrders(db);
}

export function updateBinanceOrder(requestId, updates) {
  const db = loadBinanceOrders();
  const network = config.AUTO_TRADING.USE_TESTNET ? "testnet" : "prod";
  if (!db[network]) db[network] = [];
  const idx = db[network].findIndex((o) => o.requestId === requestId);
  if (idx !== -1) {
    db[network][idx] = {
      ...db[network][idx],
      ...updates,
      updatedAt: new Date().toISOString(),
    };
    saveBinanceOrders(db);
  }
}

// ─── Binance API Actions ───
export async function initClient(testnet = true) {
  const apiKey = process.env.BINANCE_API_KEY;
  const apiSecret = process.env.BINANCE_API_SECRET;

  if (!apiKey || !apiSecret) {
    console.warn(
      "⚠️ BINANCE_API_KEY/SECRET is missing from .env. Running locally without binance node connection."
    );
  }

  // NOTE: 'binance-api-node' is a default export, requires standard call structure:
  client = Binance.default({
    apiKey,
    apiSecret,
    httpFutures: testnet
      ? "https://testnet.binancefuture.com"
      : "https://fapi.binance.com",
  });
}

function ensureClient() {
  if (!client) throw new Error("Binance API client is not initialized.");
}

export async function getAccountBalance() {
  ensureClient();
  try {
    const balances = await client.futuresAccountBalance();
    const usdtAsset = balances.find((b) => b.asset === "USDT");
    if (!usdtAsset) return 0;
    return parseFloat(usdtAsset.balance);
  } catch (err) {
    console.error("❌ Binance API Error (Balance Check):", err.message);
    throw err;
  }
}

export async function placeFuturesOrder({
  symbol,
  side,
  quantity,
  leverage,
  type = "MARKET",
}) {
  ensureClient();
  const network = config.AUTO_TRADING.USE_TESTNET ? "testnet" : "prod";
  const requestId = `${network}_${symbol}_${Date.now()}`;
  const timestamp = new Date().toISOString();

  const localState = {
    requestId,
    symbol,
    side,
    type,
    network,
    quantity,
    leverage,
    status: "NEW",
    timestamp,
  };

  try {
    await client.futuresLeverage({ symbol, leverage });
    const order = await client.futuresOrder({ symbol, side, type, quantity });

    const filledOrder = {
      ...localState,
      orderId: order.orderId,
      status: order.status || "FILLED",
      executedQty: order.executedQty,
    };
    saveBinanceOrder(filledOrder);
    console.log(
      `[${network.toUpperCase()}_MODE] Binance Futures Order Sent: ${side} ${symbol}. Status: ${
        filledOrder.status
      }`
    );
    return filledOrder;
  } catch (err) {
    localState.status = "REJECTED";
    localState.error = err.message;
    saveBinanceOrder(localState);
    console.error(`❌ Binance Futures Order Error (${symbol}):`, err.message);
    throw err;
  }
}

export async function setTakeProfitStopLoss({
  symbol,
  side,
  tpPrice,
  slPrice,
  quantity,
  parentOrderId,
}) {
  ensureClient();
  const network = config.AUTO_TRADING.USE_TESTNET ? "testnet" : "prod";
  const oppositeSide = side === "BUY" ? "SELL" : "BUY";

  const tpReqId = `${network}_tp_${symbol}_${Date.now()}`;
  const slReqId = `${network}_sl_${symbol}_${Date.now()}`;

  const localTp = {
    requestId: tpReqId,
    parentOrderId,
    symbol,
    side: oppositeSide,
    type: "TAKE_PROFIT_MARKET",
    stopPrice: tpPrice,
    status: "NEW",
    network,
  };
  const localSl = {
    requestId: slReqId,
    parentOrderId,
    symbol,
    side: oppositeSide,
    type: "STOP_MARKET",
    stopPrice: slPrice,
    status: "NEW",
    network,
  };

  try {
    const [tpRes, slRes] = await Promise.all([
      client.futuresOrder({
        symbol,
        side: oppositeSide,
        type: "TAKE_PROFIT_MARKET",
        stopPrice: tpPrice,
        closePosition: "true",
      }),
      client.futuresOrder({
        symbol,
        side: oppositeSide,
        type: "STOP_MARKET",
        stopPrice: slPrice,
        closePosition: "true",
      }),
    ]);

    localTp.orderId = tpRes.orderId;
    localTp.status = tpRes.status || "NEW";
    localSl.orderId = slRes.orderId;
    localSl.status = slRes.status || "NEW";
    saveBinanceOrder(localTp);
    saveBinanceOrder(localSl);

    console.log(
      `[${network.toUpperCase()}_MODE] TP & SL Traps Sent -> TP_ID: ${
        tpRes.orderId
      }, SL_ID: ${slRes.orderId}`
    );
    return { tp: localTp, sl: localSl };
  } catch (err) {
    console.error(`❌ TP/SL Mount Failure (${symbol}):`, err.message);
    throw err;
  }
}

export async function cancelAllOrders(symbol) {
  ensureClient();

  try {
    await client.futuresCancelAll({ symbol });

    // Sync local DB
    const db = loadBinanceOrders();
    const network = config.AUTO_TRADING.USE_TESTNET ? "testnet" : "prod";
    if (db[network]) {
      db[network] = db[network].map((o) =>
        o.symbol === symbol && o.status === "NEW"
          ? { ...o, status: "CANCELED" }
          : o
      );
      saveBinanceOrders(db);
    }
    return true;
  } catch (err) {
    console.error(`❌ Error canceling orders (${symbol}):`, err.message);
    return false;
  }
}

export async function syncOrderStatus(requestId) {
  const db = loadBinanceOrders();
  const network = config.AUTO_TRADING.USE_TESTNET ? "testnet" : "prod";
  if (!db[network]) return;

  const order = db[network].find((o) => o.requestId === requestId);
  if (
    !order ||
    !order.orderId ||
    order.status === "FILLED" ||
    order.status === "CANCELED"
  )
    return;

  ensureClient();
  try {
    const apiOrder = await client.futuresOrderStatus({
      symbol: order.symbol,
      orderId: order.orderId,
    });
    if (apiOrder.status !== order.status) {
      updateBinanceOrder(requestId, { status: apiOrder.status });
      console.log(
        `[SYNC] Order ${order.orderId} status shifted: ${order.status} -> ${apiOrder.status}`
      );
    }
  } catch (err) {
    if (err.message.includes("does not exist")) {
      updateBinanceOrder(requestId, { status: "CANCELED" });
    }
  }
}
