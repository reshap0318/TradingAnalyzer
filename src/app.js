import "dotenv/config";
import express from "express";
import cors from "cors";
import path from "path";
import { fileURLToPath } from "url";
import config from "./config.js";
import { initClient as initBinanceTrader } from "./crypto/binanceTrader.js";

import cryptoRoutes from "./crypto/cryptoRoutes.js";
import sahamRoutes from "./saham/sahamRoutes.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const app = express();
app.use(cors());
app.use(express.json());

// Initialize optional executor services
if (config.AUTO_TRADING.ENABLED) {
  initBinanceTrader(config.AUTO_TRADING.USE_TESTNET).catch((e) =>
    console.log("Init node skipped:", e.message)
  );
}

const getCurrentWIB = () => {
  return new Date().toLocaleString("id-ID", {
    timeZone: "Asia/Jakarta",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
};

// Global API routes
app.use("/crypto", cryptoRoutes);
app.use("/saham", sahamRoutes);

// Optional: Global Error Handler
app.use((err, req, res, next) => {
  console.error("Global Catch Error:", err);
  res
    .status(500)
    .json({ error: "Internal Server Error", message: err.message });
});

const PORT = process.env.PORT || 3000;

app.get("/health", (req, res) =>
  res.json({ status: "ok", timestamp: getCurrentWIB() })
);

app.listen(PORT, "0.0.0.0", () => {
  console.log(`
╔══════════════════════════════════════════════════════════════╗
║            TRADING ANALYZER API - RUNNING                    ║
╠══════════════════════════════════════════════════════════════╣
║  Server: http://0.0.0.0:${PORT}                                 ║
║                                                              ║
║  SAHAM:                                                      ║
║  GET  /saham/analyze?symbol=CODE     Full analysis           ║
║  GET  /saham/signal?symbol=CODE      Quick signal            ║
║  GET  /saham/raw?symbol=CODE         Get raw data            ║
║  POST /saham/simulate/trade          Simulate trade          ║
║  GET  /saham/simulate/summary        Simulation summary      ║
║  GET  /saham/simulate/history        Simulation history      ║
║                                                              ║
║  CRYPTO:                                                     ║
║  GET  /crypto/analyze?symbol=PAIR    Full analysis           ║
║  GET  /crypto/raw?symbol=PAIR        Get raw data            ║
║  POST /crypto/simulate/trade         Simulate trade          ║
║  GET  /crypto/simulate/summary       Simulation summary      ║
║  GET  /crypto/simulate/history       Simulation history      ║
║  GET  /crypto/trade/status           Auto-trade status       ║
║  POST /crypto/trade/auto             Start auto-trade        ║
║  POST /crypto/trade/start            Start auto-trade        ║
║  POST /crypto/trade/stop             Stop auto-trade         ║
╚══════════════════════════════════════════════════════════════╝
`);

  if (config.AUTO_TRADING.ENABLED) {
    console.log(`\n⚠️ AUTO TRADING IS ENABLED`);
    console.log(
      `- Network: ${config.AUTO_TRADING.USE_TESTNET ? "TESTNET" : "LIVE"}`
    );
  }
});
