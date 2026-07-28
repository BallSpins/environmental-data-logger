# Panduan Karakterisasi Sinyal dan Metrologi Sensor AHT10

Dokumen ini mendefinisikan protokol eksperimental untuk mengarahkan riset ke ranah **Fisika-Instrumentasi**, dengan menekankan analisis karakteristik transduser, peredaman derau (*noise*), dan metrik evaluasi sinyal menggunakan sensor suhu dan kelembapan AHT10, yang didukung oleh arsitektur perangkat keras tertanam berbasis *dual-core* dan memori non-volatil eksternal.

---

## 1. Fokus Metodologi & Batasan Ruang Lingkup

Riset ini berfokus penuh pada karakteristik fisis rantai pengukuran (*measurement chain*) dan integritas data sensor:

* **Transduser:** Sensor AHT10 (I2C) yang membaca parameter suhu ($T$) dan kelembapan relatif ($RH$).
* **Signal Conditioning:** Penerapan filter digital *Exponential Moving Average* (EMA).
* **Objek Analisis:** Karakteristik derau statis, waktu tanggap (*thermal lag*), stabilitas waktu sampling, dan perbandingan metrik metrologi (*Signal-to-Noise Ratio* serta RMSE).

---

## 2. Protokol 1: Karakterisasi Derau Dasar (*Baseline Noise Characterization*)

Sebelum menerapkan filter, perilaku bawaan dari sensor AHT10 pada kondisi tunak (*steady-state*) harus diukur untuk mengetahui tingkat ketidakpastian pengukurannya.

* **Cara Melakukannya:**
1. Tempatkan sensor AHT10 di dalam wadah tertutup yang terisolasi dari fluktuasi angin atau panas ekstrem.
2. Atur mikrokontroler untuk membaca data mentah (*raw data*) AHT10 secara kontinu pada interval waktu yang presisi.
3. Simpan stempel waktu (*timestamp*) dan nilai mentah ke dalam sistem penyimpanan.
4. Masukkan data ke lingkungan analisis Python untuk menghitung parameter statistik dasar: Simpangan baku ($\sigma$), varians, dan *Noise Floor* dari transduser.



---

## 3. Protokol 2: Pengujian Respons Transien & *Phase Lag*

Filter EMA meredam derau frekuensi tinggi, namun efek samping fisiknya adalah timbulnya keterlambatan waktu (*time delay* atau *phase lag*) pada respons sensor terhadap perubahan lingkungan.

* **Cara Melakukannya:**
1. Siapkan dua kondisi lingkungan teruji yang berbeda secara mendadak (misalnya transisi suhu terkontrol).
2. Rekam respons *step input* dari sensor AHT10 saat mengalami transisi tersebut.
3. Uji beberapa variasi konstanta penghalus pada persamaan filter EMA:

$$y_t = \alpha x_t + (1-\alpha)y_{t-1}$$




(di mana $y_t$ adalah sinyal keluaran terfilter, $x_t$ adalah pembacaan mentah saat ini, dan $\alpha$ adalah koefisien bobot antara $0 < \alpha \le 1$).


4. Bandingkan respons kurva antara data mentah dan berbagai nilai $\alpha$ (misalnya $\alpha = 0.1$, $\alpha = 0.5$, $\alpha = 0.9$) untuk mengukur besar *phase lag* terhadap fenomena fisis aslinya.



---

## 4. Metrik Evaluasi Kinerja Metrologi

Keberhasilan pemfilteran sinyal pada laporan ini divalidasi menggunakan parameter metrologi kuantitatif:

* **Signal-to-Noise Ratio (SNR) Improvement:** Menghitung peningkatan rasio sinyal terhadap derau sebelum dan sesudah penerapan filter EMA pada data *steady-state*.
* **Root Mean Square Error (RMSE):** Menghitung besar deviasi antara sinyal terfilter terhadap nilai referensi rata-rata bergerak jangka panjang untuk memastikan akurasi pengukuran tidak terdegradasi berlebihan.

---

## 5. Peran Arsitektur Dual-Core dan W25Q32 Flash Memory dalam Metrologi

Arsitektur perangkat keras tidak diposisikan sekadar sebagai utilitas jaringan, melainkan sebagai elemen krusial untuk menjamin keabsahan pengukuran fisik:

* **Deterministic Sampling (Peniadaan Jitter melalui Core 0):**
* Proses akuisisi data sensor dan eksekusi algoritma EMA diisolasi secara ketat pada **Core 0**.
* Pemisahan ini memastikan frekuensi *sampling* benar-benar stabil dan bebas dari interupsi tumpukan protokol jaringan (*network stack*), sehingga interval waktu ($\Delta t$) antar-sampel konsisten dan perhitungan matematis filter tidak cacat.


* **Non-Volatile Buffer (W25Q32 Flash Memory):**
* Berfungsi sebagai penyangga lokal berkecepatan tinggi untuk mengamankan aliran data mentah maupun data terfilter saat terjadi disrupsi eksternal.
* Memastikan seluruh matriks eksperimen metrologi tetap utuh dan terekam secara sinkron sebelum dievakuasi secara asinkron oleh **Core 1**, mencegah terjadinya kehilangan data (*data drop*) selama pengujian jangka panjang.