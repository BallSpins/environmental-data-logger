# Evaluasi Kuantitatif Phase Lag Filter EMA (Protokol 2 - Teroptimasi)

## 1. Konteks Komputasi
Evaluasi ini mengintegrasikan hasil pencarian titik optimum berbasis Fungsi Biaya (*Cost Function*) kalkulus (`protocol2-calculus-optimization.md`), dijalankan independen untuk kanal suhu dan kelembapan. Komputasi dilakukan pada rentang waktu kritis pasca-injeksi termal (menit 3.0 hingga 6.0) untuk mengukur deviasi absolut antara sinyal hasil filter (EMA) dengan data mentah.

Metrik yang dihitung meliputi:
*   **Root Mean Square Error (RMSE):** Rata-rata akar kuadrat galat selama jendela transien.
*   **Max Lag Error:** Deviasi absolut terbesar (keterlambatan fase maksimal) antara kurva filter dan data aktual.

## 2. Tabel Hasil Komputasi Numerik

### 2.1 Suhu (T) — optimal $\alpha_T = 0.1494$
| Parameter ($\alpha$) | $\tau$ (detik, eksak, $\Delta t=2.051$s) | RMSE (°C) | Max Lag Error (°C) | Waktu Max Error (Menit) |
| :--- | :--- | :--- | :--- | :--- |
| **0.05** | ≈ 40.0 | 0.2192 | 0.2686 | 5.61 |
| **0.0666** | ≈ 29.8 | 0.1713 | 0.2040 | 5.61 |
| **0.1** | ≈ 19.5 | 0.1167 | 0.1585 | 3.39 |
| **0.1494** | **≈ 12.7** | **0.0771** | **0.1333** | **3.39** |
| **0.2** | ≈ 9.2 | 0.0558 | 0.1119 | 3.39 |
| **0.4** | ≈ 4.0 | 0.0226 | 0.0616 | 3.35 |

### 2.2 Kelembapan (RH) — optimal $\alpha_{RH} = 0.2260$
| Parameter ($\alpha$) | $\tau$ (detik, eksak, $\Delta t=2.051$s) | RMSE (%RH) | Max Lag Error (%RH) | Waktu Max Error (Menit) |
| :--- | :--- | :--- | :--- | :--- |
| **0.05** | ≈ 40.0 | 3.3247 | 5.0128 | 5.13 |
| **0.0666** | ≈ 29.8 | 2.6110 | 4.0079 | 5.13 |
| **0.1** | ≈ 19.5 | 1.7840 | 3.4973 | 3.69 |
| **0.2** | ≈ 9.2 | 0.8657 | 2.5745 | 3.69 |
| **0.2260** | **≈ 8.0** | **0.7560** | **2.3984** | **3.69** |
| **0.4** | ≈ 4.0 | 0.3815 | 1.5178 | 3.69 |

## 3. Interpretasi Metrologi

**A. Verifikasi Titik Minimum per Kanal (relatif terhadap bobot $W_1=W_2=0.5$)**
Untuk suhu, Cost Function menempatkan optimal di $\alpha_T = 0.1494$, menurunkan RMSE ke **0.0771°C** dan Max Lag Error ke **0.1333°C**. Untuk kelembapan, optimal berada di $\alpha_{RH} = 0.2260$, menurunkan RMSE ke **0.7560 %RH** dan Max Lag Error ke **2.3984 %RH** — keduanya lebih baik dibanding parameter heuristik awal ($\alpha = 0.1$, yang semula dipakai untuk kedua kanal).

**B. Konsistensi Waktu Respons (*Temporal Locking*)**
Pada filter lambat ($\alpha=0.05, 0.0666$), galat terbesar pada kedua kanal terjadi di penghujung jendela observasi (menit 5.13–5.61) — algoritma gagal mengejar laju perubahan fisis. Mulai dari $\alpha=0.1$ ke atas, titik galat maksimal konsisten "terkunci" lebih awal (menit 3.39 untuk T, 3.69 untuk RH), mengindikasikan karakteristik alami *low-pass filter* yang menahan kejutan sesaat (*shock response*) di detik-detik awal sebelum beradaptasi.

**C. Validasi Silang $\tau_{RH}$ terhadap Datasheet**
Nilai $\tau_{RH} \approx 8.0$ detik hasil optimasi kalkulus murni berbasis data **hampir persis sama** dengan *typical response time* $t_{63\%}=8$ detik yang dispesifikasikan datasheet AHT10 untuk kanal RH (Table 1). Ini adalah validasi silang independen yang kuat terhadap hasil optimasi.

**D. Kesimpulan Optimasi**
Parameter $\alpha_T = 0.1494$ dan $\alpha_{RH} = 0.2260$ masing-masing menjadi titik keseimbangan terbaik untuk bobot penalti simetris yang didefinisikan pada studi ini, per kanal. Keduanya menahan deviasi terburuk secara efektif, memangkas RMSE keseluruhan, dan selaras dengan *time constant* fisis intrinsik perangkat keras masing-masing kanal ($\tau_T \approx 12.7$s dalam rentang 5–30s respons suhu; $\tau_{RH} \approx 8.0$s, hampir identik dengan *typical* respons RH datasheet) tanpa mengorbankan integritas peredaman derau.