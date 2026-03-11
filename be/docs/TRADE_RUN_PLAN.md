
### ⏱️ CRON JOB UTAMA (Jalan Tiap 1 Menit)

* Tarik semua data `Trade` dari DB yang statusnya **"ACTIVE"**.
* Lakukan *looping*: Untuk setiap `Trade`, jalankan fungsi **`ProcessTrade(trade_id)`**.

---

### ⚙️ FUNGSI: `ProcessTrade(trade_id)`

#### 0️⃣ PERSIAPAN DATA

* **Tarik DB:** Get data `Trade` beserta semua relasi `TradeEntry`-nya.
* **Validasi:** Jika status Trade bukan "ACTIVE", langsung **RETURN** (stop proses).
* **Tarik Binance:** Hit API Binance untuk *GET All Open Orders* & *Recent Trades* (Simpan di *memory* sebagai *cache* buat dicek di bawah).

#### 1️⃣ FASE 1: CEK TP / SL (Prioritas Utama Pencegah Ghost Order)

* **Cek DB:** Apakah `tp_order_id` sudah ada isinya?
* **JIKA ADA:**
* Cocokkan ID TP/SL tersebut dengan data *cache* Binance.
* **Jika Status Binance = "FILLED":**
1. Update DB `Trade.status` = **"TP_HIT"** (atau "SL_HIT").
2. Update DB `Trade.closed_at` = waktu sekarang.
3. 🚨 **CRITICAL:** Hit API Binance untuk **CANCEL SEMUA** order jaring (*Entry*) yang masih ngantre ("NEW") untuk *symbol* ini.
4. Update DB status *Entry* yang di-cancel tadi jadi **"CANCELLED"**.
5. 🛑 **RETURN** (Fungsi selesai, *trade* udah *close* bawa profit/loss).


#### 2️⃣ FASE 2: SINKRONISASI JARING / ENTRY

*(Looping untuk setiap `Entry` yang ada di dalam `Trade` ini)*

* Cocokkan ID `Entry` dengan data *cache* Binance.
* **KONDISI A: Status Binance = "NEW" (Masih Ngantre)**
* Cek usia order vs `hour_expired_config`.
* *Jika Expired:* Hit API Binance **CANCEL** order $\rightarrow$ Update DB status = **"CANCELLED"**.
* *Jika Belum Expired:* Abaikan, lanjut ke *entry* berikutnya.

* **KONDISI B: Status Binance = "PARTIALLY_FILLED"**
* Update DB status = **"PARTIALLY_FILLED"**.
* Update DB `filled_qty` sesuai data Binance yang baru.

* **KONDISI C: Status Binance = "FILLED" (Dapet Barang)**
* Update DB status = **"FILLED"**.
* Update DB `filled_qty`, `filled_price`, dan `filled_at`.
* **[Logic Pasang/Update TP & SL]**
* *Skenario 1 (Ini Entry Pertama):* Jika DB belum punya `tp_order_id`
$\rightarrow$ Hit API Binance **CREATE** order TP & SL sesuai *qty* yang didapat.
$\rightarrow$ Simpan ID TP/SL barunya ke DB.
* *Skenario 2 (Ini Entry Averaging):* Jika DB sudah punya `tp_order_id`
1. Hitung `TotalQtyBaru` (Jumlah semua koin dari *entry* yang *Filled/Partially*).
2. Hit API Binance **CANCEL** order TP/SL lama (Gratis).
3. Hit API Binance **CREATE** order TP/SL baru dengan `TotalQtyBaru` (Harga target TP/SL tetap).
4. Update DB timpa ID TP/SL lama dengan ID yang baru.

#### 3️⃣ FASE 3: NETTING & FINALISASI

* **Refresh Data:** Tarik ulang data `TradeEntry` dari DB untuk dapet *state* paling *update* setelah Fase 2.
* **Kalkulasi Induk (Update ke tabel `Trade`):**
* `TotalQty` = Jumlahkan semua `filled_qty`.
* `CapitalUsed` = Jumlahkan (`filled_qty` $\times$ `filled_price`).
* `AvgEntryPrice` = `CapitalUsed` / `TotalQty`.


* **Evaluasi Dead Signal:**
* Cek apakah **SEMUA** `Entry` berstatus "CANCELLED" atau "REJECTED".
* *Jika YA:* Update DB `Trade.status` = **"CANCELLED"** dan `closed_at` = waktu sekarang. (Artinya sinyal ini hangus karena koin terbang duluan dan gak ada satupun jaring lu yang nyangkut).