import express from "express";
import {
  analyzeSaham,
  simulateTrade,
  rawData,
  getSignalBasic,
  getSummaryStats,
  getHistoryLogs,
} from "./sahamController.js";

const router = express.Router();

router.get("/analyze", analyzeSaham);
router.get("/raw", rawData);
router.get("/signal", getSignalBasic);
router.post("/simulate/trade", simulateTrade);
router.get("/simulate/summary", getSummaryStats);
router.get("/simulate/history", getHistoryLogs);

export default router;
