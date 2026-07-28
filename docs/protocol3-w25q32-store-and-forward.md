# Dokumentasi Pengujian Protokol 3: Toleransi Kegagalan (Store-and-Forward) menggunakan W25Q32

## 1. Tujuan Pengujian
Pengujian ini bertujuan untuk memvalidasi arsitektur toleransi kegagalan (*fault tolerance*) pada *edge node* berbasis ESP32. Protokol ini menguji keandalan memori *flash* non-volatil (W25Q32) sebagai *buffer* fisik sementara (*Store-and-Forward*) untuk menggaransi integritas dan ketersediaan data (target 100% *recovery rate*) saat terjadi pemutusan koneksi jaringan secara tiba-tiba (*network disruption*).

## 2. Arsitektur dan Logika Operasi
Sistem beroperasi menggunakan isolasi tugas pada arsitektur *dual-core* ESP32 untuk memastikan kegagalan jaringan tidak mengganggu determinisme interval *sampling* sensor:

*   **Core 1 (Data Acquisition & DSP):** Secara konstan membaca data AHT10 setiap 2 detik dan menerapkan filter EMA ($\alpha = 0.1494$). Data yang telah difilter dikirim ke antrean memori internal (*thread-safe queue/buffer*).
*   **Core 0 (Network & Storage Management):** Bertugas mengambil data dari antrean dan menangani transmisi.
    *   **State Normal:** Jika koneksi MQTT terhubung, Core 0 memublikasikan (*publish*) data langsung ke *broker*.
    *   **State Disruption:** Jika MQTT terputus, Core 0 mengubah rute data (*routing*) dan melakukan operasi *Write* secara sekuensial ke dalam memori W25Q32 (beroperasi sebagai antrean FIFO/LIFO logis).
    *   **State Recovery:** Saat koneksi MQTT pulih, Core 0 melakukan operasi *Read & Flush*. Data historis yang tertahan di W25Q32 dievakuasi (*forward*) ke *broker* sebelum sistem kembali ke *State Normal*.

## 3. Metodologi Skenario Stress Test (Disruption Test)
Pengujian dilakukan secara fisik dengan menyimulasikan pemutusan jaringan di lingkungan nyata.

1.  **Fase 1: Transmisi Normal (Menit 0 – 3)**
    *   Jalankan sistem secara normal.
    *   Pastikan data AHT10 masuk secara *real-time* ke sistem basis data di *backend*.
2.  **Fase 2: Pemutusan Koneksi Terkontrol (Menit 3 – 8)**
    *   Matikan catu daya *router* Wi-Fi secara fisik.
    *   Biarkan *edge node* (ESP32) tetap menyala dan mengakuisisi data selama 5 menit tanpa koneksi internet.
    *   Dalam fase ini, LED indikator (jika ada) atau log serial harus menunjukkan bahwa operasi *Write* ke W25Q32 sedang berlangsung.
3.  **Fase 3: Pemulihan Jaringan (Menit 8 – 15)**
    *   Nyalakan kembali *router* Wi-Fi.
    *   Observasi proses *reconnect* dari ESP32 ke *broker* MQTT.
    *   Sistem harus secara otomatis melakukan *dump/flush* seluruh data yang ada di dalam W25Q32 ke *backend*.

## 4. Parameter Evaluasi Kinerja (Metrik)
Kinerja mekanisme *Store-and-Forward* divalidasi secara kuantitatif melalui rekonsiliasi data antara *edge node* dan *backend*.

1.  **Total Data Terbentuk ($N_{\text{gen}}$):** 
    Jumlah baris data yang dihasilkan oleh Core 0 selama Fase 2. (Pada interval 2 detik selama 5 menit, target teoretis = 150 baris data).
2.  **Total Data Masuk di Backend ($N_{\text{rec}}$):** 
    Jumlah baris data historis dari rentang waktu pemutusan yang berhasil masuk ke *database* setelah fase pemulihan.
3.  **Rasio Pemulihan (Recovery Rate):**
    $$\text{Recovery Rate} = \left( \frac{N_{\text{rec}}}{N_{\text{gen}}} \right) \times 100\%$$
    *Indikator Keberhasilan:* Sistem dinyatakan lolos pengujian metrologi jika *Recovery Rate* mencapai persis **100%** (tidak ada satu pun paket data/baris (*packet loss*) yang hilang akibat interupsi transmisi).

## 5. Kesimpulan Teknis (Placeholder)
*(Bagian ini akan diisi setelah eksperimen dieksekusi. Merangkum apakah cip W25Q32 mampu menangani operasi baca/tulis sinkron tanpa kehilangan data, dan mengevaluasi seberapa cepat waktu yang dibutuhkan sistem untuk mem-flush antrean data saat jaringan kembali online).*