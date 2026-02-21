import express from "express";
import {
  analyzeCrypto,
  tradeStatus,
  tradeStop,
  tradeAuto,
  simulateTrade,
  rawData,
  getSummaryStats,
  getHistoryLogs,
} from "./cryptoController.js";

const router = express.Router();

router.get("/analyze", analyzeCrypto);
router.get("/trade/status", tradeStatus);
router.post("/trade/stop", tradeStop);
router.post("/trade/auto", tradeAuto);
router.get("/raw", rawData);
router.post("/simulate/trade", simulateTrade);
router.get("/simulate/summary", getSummaryStats);
router.get("/simulate/history", getHistoryLogs);

export default router;
