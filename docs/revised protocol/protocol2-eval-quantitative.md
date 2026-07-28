# Evaluasi Kuantitatif Phase Lag Filter EMA (Protokol 2)

## 1. Konteks Komputasi
Evaluasi ini merupakan suplemen kuantitatif untuk melengkapi observasi visual pada pengujian respons transien (Protokol 2). Komputasi dilakukan pada rentang waktu kritis pasca-injeksi termal (menit 3.0 hingga 6.0) untuk mengukur deviasi absolut antara sinyal hasil filter (EMA) dengan data mentah, untuk kanal suhu dan kelembapan.

Metrik yang dihitung meliputi:
*   **Root Mean Square Error (RMSE):** Rata-rata akar kuadrat galat selama jendela transien.
*   **Max Lag Error:** Deviasi absolut terbesar (keterlambatan fase maksimal) antara kurva filter dan data aktual.

## 2. Tabel Hasil Komputasi Numerik

### 2.1 Suhu (T)
| Parameter ($\alpha$) | $\tau$ (detik, eksak) | RMSE (°C) | Max Lag Error (°C) | Waktu Max Error (Menit) |
| :--- | :--- | :--- | :--- | :--- |
| **0.05** | ≈ 40.0 | 0.2192 | 0.2686 | 5.61 |
| **0.0666** | ≈ 29.8 | 0.1713 | 0.2040 | 5.61 |
| **0.1** | **≈ 19.5** | **0.1167** | **0.1585** | **3.39** |
| **0.2** | ≈ 9.2 | 0.0558 | 0.1119 | 3.39 |
| **0.4** | ≈ 4.0 | 0.0226 | 0.0616 | 3.35 |

### 2.2 Kelembapan (RH)
| Parameter ($\alpha$) | $\tau$ (detik, eksak) | RMSE (%RH) | Max Lag Error (%RH) | Waktu Max Error (Menit) |
| :--- | :--- | :--- | :--- | :--- |
| **0.05** | ≈ 40.0 | 3.3247 | 5.0128 | 5.13 |
| **0.0666** | ≈ 29.8 | 2.6110 | 4.0079 | 5.13 |
| **0.1** | **≈ 19.5** | **1.7840** | **3.4973** | **3.69** |
| **0.2** | ≈ 9.2 | 0.8657 | 2.5745 | 3.69 |
| **0.4** | ≈ 4.0 | 0.3815 | 1.5178 | 3.69 |

Catatan: nilai RMSE/Max Lag Error kanal RH jauh lebih besar secara nominal dibanding kanal T — ini konsekuensi skala fisis kejadian (transien RH ≈26 %RH vs. transien suhu ≈1.5°C), bukan indikasi kanal RH lebih berisik.

## 3. Interpretasi Metrologi

**A. Verifikasi Besaran Phase Lag Maksimal**
Filter dengan redaman agresif ($\alpha = 0.05$) menghasilkan deviasi absolut maksimum hingga 0.2686°C (T) / 5.0128 %RH (RH) terhadap data aktual, keduanya pada bagian akhir jendela transien.

**B. Posisi Keseimbangan Parameter $\alpha = 0.1$ (heuristik, awalnya hanya untuk T)**
Untuk suhu, $\alpha = 0.1$ menahan *Max Lag Error* pada 0.1585°C, RMSE 0.1167°C, reduksi standar deviasi teoretis ≈77.1%. Untuk RH pada $\alpha$ yang sama, *Max Lag Error* masih 3.4973 %RH — jauh lebih longgar dibanding hasil optimasi RH yang sesungguhnya ($\alpha_{RH}=0.2260$, lihat dokumen *-optimized*), mengindikasikan bahwa $\alpha=0.1$ bukan titik seimbang yang tepat untuk kanal RH.

**C. Pergeseran Momentum Galat Terbesar**
Pada filter lambat ($\alpha=0.05, 0.0666$), galat terbesar pada kedua kanal terjadi di penghujung jendela observasi (menit 5.13–5.61) — konstanta waktu algoritmik terlalu lambat mengejar laju perubahan fisis. Pada $\alpha \geq 0.1$, galat terbesar terjadi lebih awal (menit 3.39 untuk T, 3.69 untuk RH) dengan adaptasi cepat sesudahnya.

**→ Lihat `protocol2-eval-quantitative-optimized.md` untuk hasil setelah $\alpha_T=0.1494$ dan $\alpha_{RH}=0.2260$ dimasukkan ke tabel pembanding.**