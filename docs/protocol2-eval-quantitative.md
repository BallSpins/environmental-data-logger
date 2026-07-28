# Evaluasi Kuantitatif Phase Lag Filter EMA (Protokol 2)

## 1. Konteks Komputasi
Evaluasi ini merupakan suplemen kuantitatif untuk melengkapi observasi visual pada pengujian respons transien (Protokol 2). Komputasi dilakukan pada rentang waktu kritis pasca-injeksi termal (menit 3.0 hingga 6.0) untuk mengukur deviasi absolut antara sinyal hasil filter (EMA) dengan data suhu mentah (*raw data*). 

Metrik yang dihitung meliputi:
*   **Root Mean Square Error (RMSE):** Rata-rata akar kuadrat galat selama jendela transien.
*   **Max Lag Error:** Deviasi absolut terbesar (keterlambatan fase maksimal) antara kurva filter dan data aktual.

## 2. Tabel Hasil Komputasi Numerik

| Parameter ($\alpha$) | Toleransi Lag ($\tau_{\text{max}}$) | RMSE (°C) | Max Lag Error (°C) | Waktu Max Error Terjadi (Menit) |
| :--- | :--- | :--- | :--- | :--- |
| **0.05** | $\approx$ 40 detik | 0.2192 | 0.2686 | 5.61 |
| **0.0666** | $\approx$ 30 detik | 0.1713 | 0.2040 | 5.61 |
| **0.1** | **$\approx$ 20 detik** | **0.1167** | **0.1585** | **3.39** |
| **0.2** | $\approx$ 10 detik | 0.0558 | 0.1119 | 3.39 |
| **0.4** | $\approx$ 5 detik | 0.0226 | 0.0616 | 3.35 |

## 3. Interpretasi Metrologi

Data numerik di atas memberikan landasan matematis yang objektif untuk mengevaluasi kinerja filter:

**A. Verifikasi Besaran Phase Lag Maksimal**
Filter dengan redaman agresif ($\alpha = 0.05$) menghasilkan deviasi absolut maksimum hingga 0.2686°C terhadap data aktual pada menit ke-5.61. Untuk instrumentasi observasi suhu, galat algoritmik yang melampaui 0.25°C tidak dapat ditoleransi karena akan mendegradasi akurasi fisis pembacaan secara signifikan.

**B. Posisi Keseimbangan Parameter $\alpha = 0.1$**
Nilai $\alpha = 0.1$ terbukti mampu menahan penyimpangan transien (*Max Lag Error*) pada batas 0.1585°C, dengan total RMSE selama fase transisi tertahan di angka 0.1167°C. Metrik ini mengonfirmasi bahwa $\alpha = 0.1$ secara efektif mencegah galat *phase lag* menembus ambang batas 0.2°C, namun tetap memberikan rasio redaman derau yang memadai (≈77% secara teoretis) dibandingkan parameter yang lebih besar (seperti $\alpha = 0.4$).

**C. Pergeseran Momentum Galat Terbesar**
Distribusi temporal dari galat maksimal menunjukkan dinamika pelacakan sinyal yang berbeda:
*   Pada filter lambat ($\alpha = 0.05$ dan $0.0666$), akumulasi galat terbesar terjadi di penghujung jendela observasi (menit 5.61). Hal ini mengindikasikan bahwa konstanta waktu algoritmik terlalu lambat untuk mengejar laju perubahan fisis suhu.
*   Pada filter $\alpha = 0.1$, galat terbesar terjadi sesaat setelah *step input* (menit 3.39), namun algoritma dengan cepat beradaptasi dan memperkecil deviasinya. Ini membuktikan bahwa traksi matematis dari $\alpha = 0.1$ memiliki keselarasan yang lebih baik terhadap *time constant* fisis sistem termal lingkungan.