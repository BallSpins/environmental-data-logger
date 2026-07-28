# Blueprint Eksperimen dan Laporan Instrumentasi

## 1. Fokus Penelitian (Scope)
Penelitian ini berfokus pada **satu scope utama: Integritas Data**. 

Penapisan sinyal (EMA) dan toleransi kegagalan (Store-and-Forward) diposisikan sebagai dua instrumen metodologis untuk mengamankan scope utama tersebut.
*   **Masalah Utama:** Degradasi data akibat derau fisis (kualitas buruk) dan *packet loss* akibat koneksi terputus (ketersediaan buruk).
*   **Parameter 1 (Kualitas):** Penggunaan filter *Exponential Moving Average* (EMA) untuk menekan *transient noise*.
*   **Parameter 2 (Ketersediaan):** Penggunaan memori non-volatil (SD Card) sebagai *buffer* fisik untuk menjamin tidak ada data yang hilang saat *network disruption*.

## 2. Roadmap Eksekusi Eksperimen Fisik
Setelah perangkat keras (AHT10, ESP32, Modul SD Card, RTC) tersedia, lakukan protokol pengujian berikut secara berurutan:

### A. Pengujian DSP (Digital Signal Processing)
1.  Jalankan sensor dan kumpulkan *raw data* yang memiliki fluktuasi (*noise*) tanpa pemrosesan apa pun.
2.  Aktifkan algoritma EMA dengan memvariasikan bobot penghalus ($\alpha$), misalnya 0.1, 0.5, dan 0.9.
3.  Simpan hasil akuisisi. Anda memerlukan set data ini untuk membuat grafik *overlay* (raw data vs filtered data) guna menganalisis *time-delay* dan efektivitas peredaman derau.

### B. Pengujian Fault Tolerance (Ketersediaan Data)
1.  Jalankan sistem dalam kondisi jaringan normal (terhubung ke broker MQTT).
2.  Matikan *router* secara fisik selama durasi tertentu (misal: 5-10 menit) di tengah siklus akuisisi.
3.  Hidupkan kembali *router*.
4.  Verifikasi rasio pemulihan data: Bandingkan jumlah baris log yang masuk ke SD Card dengan jumlah baris yang berhasil diterima oleh *backend database*. Target keberhasilan sistem adalah 100% *recovery rate*.

---

## 3. Kerangka Dasar Laporan LaTeX (main.tex)

**Abstrak**
Maksimal 250 kata. Definisikan masalah integritas data di *edge node*. Sebutkan metode yang digunakan (EMA dan Store-and-Forward). Tutup dengan matriks keberhasilan fisis dan persentase pemulihan transmisi.

**1. Pendahuluan**
*   **Konteks:** Kebutuhan data representatif pada pemantauan lingkungan. Sensor standar menghasilkan sinyal dengan *noise* tinggi.
*   **Masalah:** Data kotor merusak komputasi *backend*; transmisi *real-time* tunggal rentan terhadap *packet loss*.
*   **Tujuan:** Memvalidasi arsitektur akuisisi mandiri yang mampu membersihkan data secara on-board dan menggaransi ketersediaan data saat *disruption*.

**2. Desain dan Metodologi Eksperimen**
Hindari penulisan kode program mentah. Gunakan pendekatan matematis dan logika sistem.
*   **2.1 Penapisan Sinyal Digital (DSP):** Definisikan pemodelan filter menggunakan persamaan matematis:
    $$y_t = \alpha x_t + (1-\alpha)y_{t-1}$$
    Keterangan: $y_t$ (sinyal keluaran), $x_t$ (sinyal masukan saat ini), dan $\alpha$ (konstanta penghalus).
*   **2.2 Arsitektur Toleransi Kegagalan (Store-and-Forward):** Buat deskripsi logika sirkuit data. Jika MQTT putus $\rightarrow$ Operasi *Write* ke `datalog.bin`. Jika MQTT pulih $\rightarrow$ Operasi *Read & Flush* dari FIFO ke *broker*.

**3. Hasil dan Pembahasan**
Bagian inti dari pembuktian eksperimen fisik.
*   **3.1 Analisis Kinerja Filter EMA:** Tampilkan grafik komparasi waktu terhadap suhu. Buat *overlay* antara raw data dan filtered data. Analisis kompromi (*trade-off*) dari nilai $\alpha$ terhadap respons sensor.
*   **3.2 Uji Keandalan Transmisi (Disruption Test):** Jabarkan skenario *stress test* konektivitas. Gunakan tabel untuk mengkomparasi rasio data yang diakuisisi secara luring versus data yang berhasil dievakuasi ke *cloud*.

**4. Kesimpulan**
Satu paragraf padat. Validasi secara empiris bahwa integrasi pemrosesan EMA dan mekanisme *Store-and-Forward* terbukti mengamankan integritas fisis dan logis pada instrumen akuisisi berbiaya rendah.