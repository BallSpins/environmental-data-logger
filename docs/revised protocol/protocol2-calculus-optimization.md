# Optimasi Parameter Filter EMA Berbasis Fungsi Biaya (Cost Function)

## 1. Latar Belakang dan Rasionalisasi Matematis
Penentuan nilai $\alpha$ pada iterasi sebelumnya ($\alpha = 0.1$, lihat `protocol2-aht10-ema-transient.md`) masih didasarkan pada metode heuristik dan inspeksi visual terhadap kompromi antara redaman derau (Protokol 1) dan *phase lag* (Protokol 2), dan hanya dievaluasi untuk kanal suhu. Pendekatan tersebut memiliki kelemahan metodologis karena menyisakan ruang bagi bias observasi manusia, dan tidak memperhitungkan bahwa kanal RH punya karakteristik derau dan dinamika transien yang berbeda dari kanal suhu.

Untuk mengurangi ketergantungan pada inspeksi visual, masalah ini direduksi menjadi persoalan optimasi matematis menggunakan Fungsi Biaya (*Cost Function*), yang dijalankan secara independen untuk tiap kanal. Fungsi objektif $J(\alpha)$ dirancang untuk mencari titik minimum dari akumulasi dua penalti kerugian sistem:
1.  **Penalti Derau ($C_{\text{noise}}$):** Berbanding lurus dengan $\alpha$, dievaluasi menggunakan simpangan baku dari dataset kondisi tunak (Protokol 1), per kanal.
2.  **Penalti Phase Lag ($C_{\text{lag}}$):** Berbanding terbalik dengan $\alpha$, dievaluasi menggunakan *Root Mean Square Error* (RMSE) dari dataset transien (Protokol 2), per kanal.

Persamaan Fungsi Biaya yang dinormalisasi (identik untuk kedua kanal):
$$J(\alpha) = W_1 \cdot \hat{C}_{\text{noise}}(\alpha) + W_2 \cdot \hat{C}_{\text{lag}}(\alpha)$$
di mana $W_1 = 0.5$ dan $W_2 = 0.5$ merupakan bobot penyeimbang simetris yang **dipilih sebagai asumsi desain studi ini** untuk menjaga prioritas setara antara integritas sinyal dan kecepatan respons sistem. Bobot yang berbeda (misalnya memprioritaskan kecepatan respons untuk aplikasi yang butuh deteksi perubahan cepat) akan menggeser titik optimal ke nilai $\alpha$ yang berbeda.

## 2. Hasil Komputasi Numerik

### 2.1 Kanal Suhu (T)
Evaluasi numerik iteratif pada rentang $\alpha \in [0.01, 0.99]$ (500 titik) menghasilkan kurva $J(\alpha)$ berbentuk parabola dengan titik minimum pada:
*   **Nilai Optimal (relatif terhadap $W_1=W_2=0.5$):** $\alpha_T = 0.1494$
*   **Konstanta Waktu ($\Delta t$ terukur = 2.051 detik):**
    $$\tau_T = \frac{-\Delta t}{\ln(1-\alpha_T)} = \frac{-2.051}{\ln(0.8506)} \approx 12.67 \text{ detik}$$
*   **Reduksi standar deviasi teoretis:** $\approx 71.6\%$ ($\alpha_T/(2-\alpha_T) = 0.0808$)

### 2.2 Kanal Kelembapan (RH)
Dengan metodologi identik, evaluasi untuk kanal RH menghasilkan:
*   **Nilai Optimal (relatif terhadap $W_1=W_2=0.5$):** $\alpha_{RH} = 0.2260$
*   **Konstanta Waktu (rumus eksak, $\Delta t$ terukur = 2.051 detik):**
    $$\tau_{RH} = \frac{-\Delta t}{\ln(1-\alpha_{RH})} = \frac{-2.051}{\ln(0.7740)} \approx 8.00 \text{ detik}$$
*   **Reduksi standar deviasi teoretis:** $\approx 64.3\%$ ($\alpha_{RH}/(2-\alpha_{RH}) = 0.1274$)

Nilai $\alpha_{RH}$ jauh lebih besar dari $\alpha_T$ — Fungsi Biaya RH secara numerik "memilih" filter yang jauh lebih ringan/responsif dibanding filter suhu, karena dinamika transien kelembapan pada pengujian ini lebih cepat relatif terhadap derau tunaknya dibanding suhu.

## 3. Analisis dan Signifikansi Metrologi
1.  **Koreksi Titik Buta (*Blind Spot*):** Secara visual, kurva awal $\alpha = 0.1$ (dipakai untuk suhu) tampak meredam derau dengan baik namun menyisakan *lag* di awal fase transisi, sementara $\alpha = 0.2$ memiliki kecepatan ideal namun terbukti melewatkan derau LSB. Komputasi numerik menempatkan titik keseimbangan (untuk bobot simetris yang dipilih) persis di $\alpha_T = 0.1494$ — menunjukkan bahwa tebakan interval diskret manusia ($\alpha=0.1$ atau $0.2$) tidak sepresisi solusi numerik kontinu. Pola serupa berlaku untuk RH, di mana nilai heuristik lama ($\alpha=0.1$, dirancang untuk suhu) ternyata bukan titik optimal untuk kelembapan.
2.  **Keselarasan Dinamika Fisis — Suhu:** Konstanta waktu filter $\tau_T \approx 12.7$ detik berada dalam rentang limit respons intrinsik perangkat keras yang disebutkan datasheet AHT10 **untuk kanal suhu** — *response time* $t_{63\%}$ = **5–30 detik** (Table 3, "Temperature", kondisi *min–max*). Catatan: angka "8 detik" yang sempat dipakai pada draf sebelumnya untuk klaim ini keliru — itu adalah *typical response time* kanal **kelembapan (RH)** (Table 1), bukan suhu.
3.  **Keselarasan Dinamika Fisis — Kelembapan:** Menariknya, konstanta waktu filter RH hasil optimasi, $\tau_{RH} \approx 8.0$ detik, justru **hampir persis sama** dengan *typical response time* $t_{63\%}$ = 8 detik yang dispesifikasikan datasheet AHT10 untuk kanal RH (Table 1, "Relative humidity", kondisi *typical*). Ini adalah validasi silang yang kuat: optimasi Cost Function murni berbasis data empiris (independen dari datasheet) konvergen ke konstanta waktu yang secara independen juga tercatat sebagai spesifikasi fisis bawaan sensor.
4.  **Konsekuensi Implementasi:** Karena $\alpha_T \neq \alpha_{RH}$, implementasi akhir pada *edge node* memerlukan **dua filter EMA independen** — satu per kanal — bukan satu filter tunggal yang dipakai untuk kedua besaran.

## 4. Kesimpulan
Nilai $\alpha_T = 0.1494$ (suhu) dan $\alpha_{RH} = 0.2260$ (kelembapan) ditetapkan sebagai parameter operasional final untuk filter *Exponential Moving Average* pada instrumen *edge node* AHT10, **di bawah asumsi bobot penalti simetris ($W_1=W_2=0.5$) yang didefinisikan pada studi ini**. Ketetapan ini tidak lagi bergantung pada interpretasi komparatif visual, melainkan pada minimisasi numerik Fungsi Biaya yang dapat direplikasi dan dipertanggungjawabkan secara objektif per kanal — dengan catatan bahwa perubahan prioritas aplikasi (bobot $W_1, W_2$ berbeda) akan menghasilkan titik optimal $\alpha$ yang berbeda pula.