import express from "express";
import {
  analyzeSaham,
  simulateTrade,
  rawData,
  getSignalBasic,
  getSummaryStats,
  getHistoryLogs,
  getActiveLogs,
} from "./sahamController.js";

const router = express.Router();

router.get("/analyze", analyzeSaham);
router.get("/raw", rawData);
router.get("/signal", getSignalBasic);
router.post("/simulate/trade", simulateTrade);
router.get("/simulate/summary", getSummaryStats);
router.get("/simulate/history", getHistoryLogs);
router.get("/simulate/active", getActiveLogs);

export default router;
