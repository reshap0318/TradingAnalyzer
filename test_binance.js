import "dotenv/config";
import Binance from "binance-api-node";

async function testConnection() {
  console.log("Menghubungkan ke Binance Testnet...");

  const client = Binance.default({
    apiKey: process.env.BINANCE_API_KEY,
    apiSecret: process.env.BINANCE_API_SECRET,
    httpFutures: "https://testnet.binancefuture.com",
  });

  try {
    // Ping server
    await client.futuresPing();
    console.log("✅ Ping server berhasil!");

    // Cek saldo Futures
    const accountInfo = await client.futuresAccountInfo();
    console.log("✅ Login berhasil!");
    console.log("-----------------------------------------");
    console.log(
      `Batas Margin Anda (Cross): $${accountInfo.totalCrossWalletBalance}`
    );
    console.log(`Balance Tersedia: $${accountInfo.availableBalance}`);
    console.log("-----------------------------------------");
    console.log(
      "API Key Testnet Anda VALID dan siap digunakan untuk Auto-Trading!"
    );
  } catch (error) {
    console.error("❌ Gagal terhubung ke Testnet:");
    console.error(error.message);
    if (error.message.includes("API-key format invalid")) {
      console.error(
        "-> Pastikan Anda sudah copy-paste API Key dengan benar di .env"
      );
    } else if (
      error.message.includes("Signature for this request is not valid")
    ) {
      console.error("-> Secret Key Anda salah atau tidak cocok dengan API Key");
    }
  }
}

testConnection();
