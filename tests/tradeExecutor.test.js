import test from "node:test";
import assert from "node:assert/strict";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
import { checkSafetyRules } from "../src/shared/tradeExecutor.js";
import config from "../src/config.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const SUMMARY_FILE = path.join(__dirname, "../data/summary.json");

test("Trade Executor Safety Rules", async (t) => {
  // Backup configurations
  const backupEnabled = config.AUTO_TRADING.ENABLED;
  const backupDrawdown = config.AUTO_TRADING.SAFETY.MAX_DRAWDOWN_PERCENT;

  // Mock Summary JSON to control drawdown states
  const mockSummary = (pnlDollar) => {
    // Override getSummary behavior explicitly for this test by stubbing require or just writing forced json
    const dir = path.dirname(SUMMARY_FILE);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });

    // Write out the file so `loadJson` inside `updateSummaryData` reads it
    // Wait, updateSummaryData recalculates from HISTORY_FILE!
    // We must mock the history file as well if we want totalPnlDollar to stick.
    const HISTORY_FILE = path.join(__dirname, "../data/history.json");
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
    fs.writeFileSync(
      HISTORY_FILE,
      JSON.stringify([
        { assetType: "CRYPTO", signal: "SELL", outcome: "CLOSED", pnlDollar },
      ])
    );
  };

  t.afterEach(() => {
    config.AUTO_TRADING.ENABLED = backupEnabled;
    config.AUTO_TRADING.SAFETY.MAX_DRAWDOWN_PERCENT = backupDrawdown;
    const HISTORY_FILE = path.join(__dirname, "../data/history.json");
    if (fs.existsSync(SUMMARY_FILE)) fs.unlinkSync(SUMMARY_FILE);
    if (fs.existsSync(HISTORY_FILE)) fs.unlinkSync(HISTORY_FILE);
  });
  const validAnalysis = {
    quote: { symbol: "BTCUSDT" },
    signalResult: { signal: "BUY", confidence: 80 },
    moneyMgmt: { rekomendasiTrade: { valid: true } },
    tpsl: { riskReward: { tp: 2.0 } },
    btcMarket: { isCrash: false },
  };

  await t.test("should reject if AUTO_TRADING globally disabled", () => {
    config.AUTO_TRADING.ENABLED = false;
    mockSummary(0);
    const result = checkSafetyRules(validAnalysis);
    assert.strictEqual(result.safe, false);
    assert.match(result.reasons[0], /Auto-Trading is globally disabled/);
  });

  await t.test("should reject if Signal is WAIT", () => {
    config.AUTO_TRADING.ENABLED = true;
    mockSummary(0);
    const result = checkSafetyRules({
      ...validAnalysis,
      signalResult: { signal: "WAIT" },
    });
    assert.strictEqual(result.safe, false);
    assert.match(result.reasons.join(","), /Signal is WAIT/);
  });

  await t.test("should reject if Maximum Drawdown is breached", () => {
    config.AUTO_TRADING.ENABLED = true;
    config.AUTO_TRADING.SAFETY.MAX_DRAWDOWN_PERCENT = 10;

    // Create a 15% simulated loss dynamically regardless of what config is
    const dynamicLoser = ((config.CRYPTO.DEFAULT_CAPITAL * 15) / 100) * -1;
    mockSummary(dynamicLoser);
    const result = checkSafetyRules(validAnalysis);

    assert.strictEqual(result.safe, false);
    assert.match(result.reasons.join(","), /Maximum Drawdown Limit Hit/);
  });

  await t.test("should pass valid trades under healthy environments", () => {
    config.AUTO_TRADING.ENABLED = true;
    // Positive Trade Environment
    mockSummary(10);

    const result = checkSafetyRules(validAnalysis);
    assert.strictEqual(result.safe, true, "Healthy trade was falsely blocked");
    assert.strictEqual(result.reasons.length, 0);
  });
});
