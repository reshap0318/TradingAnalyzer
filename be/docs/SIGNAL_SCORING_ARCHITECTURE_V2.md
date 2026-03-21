# Proposal Peningkatan Arsitektur Skoring Sinyal (V2: Driver & Filter Model)

## 1. Latar Belakang & Masalah Bisnis (POV Trading)
Dalam operasional *trading* otomatis sejauh ini, ditemukan sebuah anomali performa di mana bot sangat jarang (bahkan hampir tidak pernah) menghasilkan *Strong Signal* (Tingkat Keyakinan / *Confidence* >= 70). 

Secara bisnis, hal ini merugikan karena:
1. **Hilangnya Peluang Emas (*Opportunity Cost*)**: Bot sering kali kehilangan momen untuk masuk dengan *size* maksimal pada saat tren pasar sedang sangat kuat memihak kita (misalnya saat terjadi *breakout* harga).
2. **Terlalu Lolosnya Sinyal Lemah**: Sistem bekerja lebih banyak pada jajaran sinyal medioker/lemah yang berisiko tinggi terkena *whipsaw* (harga berbalik arah secara tipis).

Akar masalah matematis dari hal ini disebut sebagai **Sistem Saling Menetralkan (*Cancel-Out Effect*)**. Di arsitektur V1, seluruh indikator teknikal (baik itu pengikut tren maupun *oscillator*) dijumlahkan skornya secara pukul rata (Demokrasi Total). Padahal secara natural, saat indikator Tren berteriak `STRONG BUY`, indikator Oscillator pasti akan membaca pasar sedang `Overbought` (Jenuh Beli) dan memberikan skor `SELL`. Alhasil, skor akhir yang tadinya sangat bagus (+100) ditarik turun menjadi skor rata-rata (+50) oleh indikator yang berbeda sifat.

## 2. Solusi Strategis (Arsitektur V2)
Untuk memaksimalkan *Win Rate* dan ketajaman masuk pasar, arsitektur skoring harus dirombak dari sistem **"Penjumlahan Bebas"** menjadi hierarki **"Sistem Otoritas (Driver & Filter)"**. 

Pada arsitektur V2, setiap indikator tidak lagi diperlakukan setara, melainkan dibagi ke dalam 3 Jabatan/Peran (Roles):

### A. DRIVER (Penentu Arah Utama)
- **Aktor:** Moving Average (MA), MACD.
- **Wewenang Bisnis:** Hanya indikator di kategori ini yang berhak mengeluarkan skor dasar `+100` (Buy) hingga `-100` (Sell). Jika *Driver* sepakat pasar sedang kuat naik, skor dasar dicatat di angka maksimal.

### B. FILTER (Penjaga Keamanan / Sabuk Pengaman)
- **Aktor:** RSI, Stochastic, Bollinger Bands.
- **Wewenang Bisnis:** Tidak berhak menyumbang angka. Jika DRIVER berkata pasar sedang naik kuat, FILTER hanya bertugas sebagai **Pengali (*Multiplier*) Hukuman**. 
    - Jika pasar masih aman (RSI Normal), skor DRIVER tetap utuh `100%`.
    - Jika pasar sudah terlalu bahaya/jenuh (*Overbought*), FILTER berhak menjatuhkan denda (skor dikalikan pinalti `0.0` alias DIBATALKAN).
- **Dampak Bisnis:** Modal terlindungi secara mutlak dari masuk di harga pucuk, tanpa harus mengorbankan skor tren yang bagus.

### C. BOOSTER (Pengali Hadiah / Konfirmator Ekstra)
- **Aktor:** Volume, ATR, *Candle Patterns*.
- **Wewenang Bisnis:** Bertindak sebagai pendorong keyakinan (*Multiplier* Bonus). Jika *Breakout* dikonfirmasi oleh Volume yang meledak, skor DRIVER akan dilipatgandakan (misal `1.2x`).
- **Dampak Bisnis:** Saat kondisi pasar sempurna, bot bisa memicu tingkat keyakinan *STRONG BUY* dengan sangat mudah.

## 3. Manfaat Bisnis & Proyeksi ROI (Return on Investment)
Dengan diimplementasikannya Arsitektur Skoring V2 ini, bisnis *trading* algoritma kita akan mendapatkan efisiensi berikut:
1. **Peningkatan Presisi Otomatis**: Menghilangkan fenomena bot yang "ragu-ragu" saat tren sudah jelas. Ketegasan mengambil posisi akan meningkatkan persentasi untung (*Win Rate*).
2. **Eksekusi Manajemen Risiko yang Logis**: *Stop-out* sebelum masuk! Bot tidak akan lagi membeli di area *overbought* parah karena sistem FILTER memiliki hak veto mutlak, menurunkan *Maximum Drawdown* (Kerugian Maksimal) per bulan secara drastis.
3. **Optimalisasi Strategi Tanpa Konflik**: Anda (sebagai *User/Manager*) bisa memasukkan belasan indikator sekaligus ke dalam satu strategi tanpa takut strategi tersebut akan menjadi "bodoh" karena kebanyakan filter berdebat satu sama lain. Mesin tahu siapa pimpinan (Driver) dan siapa pelayan (Filter).

## 4. Rencana Implementasi Teknis (Timeline IT)
Untuk mencapai perombakan ini, pengerjaan yang diperlukan mencakup:
1. **Database Update (Tabel `indicators`)**: Menambahkan kolom `Role` bertipe Enum (`DRIVER`, `FILTER`, `BOOSTER`). (*Estimasi: 15 Menit*)
2. **Refactoring Logika Core (`internal/bot/indicators/`)**: Mengubah *output* / hasil akhir fungsi-fungsi *Oscillator* dan *Volume* yang tadinya mengeluarkan persentase (`float64` +/-) menjadi faktor pengali / *multiplier* (misal 0.5, 1.0, 1.25). (*Estimasi: 1-2 Jam*)
3. **Refactoring Layanan Kalkulasi (`signal_analyze_service.go`)**: Mengganti rumus perhitungan akhir dari *`Sum(Weight * Score)`* menjadi *`(Sum(Driver Score) * Filter Multiplier) * Booster Multiplier`*. (*Estimasi: 1 Jam*)
4. **Penyesuaian Multi-Timeframe (Opsional)**: Menerapkan hal yang sama pada rentang waktu (Timeframe Utama sebagai *Driver*, Timeframe Besar sebagai *Filter/Approval*).

---
**Status Dokumen:** `DRAFT` (Menunggu persetujuan eksekusi perombakan dari Manager/User).
