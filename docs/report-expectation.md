# Analisis dan Strategi Pengembangan: Karakterisasi Integritas Data pada Edge Node

Dokumen ini merangkum hasil bedah kritis, ekspektasi teknis, serta strategi penyesuaian sudut pandang keilmuan (dari *software-heavy* ke *instrumentasi fisika*) untuk draf karya ilmiah/proyek dengan judul:

> **"Characterization of Data Integrity on Environmental Acquisition Edge Nodes Using EMA Filters and Non-Volatile Buffers"**
> **Penulis:** Muhammad Fauzan Faturosi
> **Afiliasi:** Departemen Fisika, Universitas Pembangunan Nasional "Veteran" Jawa Timur

---

## 1. Ekspektasi Teknis Berdasarkan Judul

### A. Fokus Metodologi & Arsitektur Sistem
* **Exponential Moving Average (EMA) Filters:** Penerapan formula matematis rekursif ($lpha \cdot x_t + (1-lpha) \cdot x_{t-1}$) untuk meredam derau (*noise*) sensor lingkungan (seperti fluktuasi ADC) dengan overhead komputasi yang rendah, sangat ideal untuk mikrokontroler sumber daya terbatas.
* **Non-Volatile Buffers:** Mekanisme antrean dan penyimpanan lokal (menggunakan EEPROM, Flash eksternal, atau sistem berkas seperti LittleFS) untuk mitigasi *data loss* saat terjadi *brownout* atau pemutusan catu daya mendadak (*sudden power loss*) sebelum data dikirimkan ke server.
* **Characterization & Data Integrity:** Pengukuran metrik performa secara ketat, mencakup tingkat reliabilitas paket, latensi pemrosesan, *buffer overflow rate*, hingga efisiensi memori (RAM/Flash) pada *edge node*.

### B. Evaluasi Ruang Lingkup (Scope) & Afiliasi Fisika
* Menjembatani kesenjangan antara pendekatan murni perangkat lunak (*software/IoT implementation*) dengan rumpun keilmuan **Fisika Instrumentasi**. 
* Pendekatan fisika menuntut validasi sensor fisik (kalibrasi, analisis error statistik, dan *signal-to-noise ratio*), bukan sekadar arsitektur jaringan atau pengkodean tingkat aplikasi.

---

## 2. Strategi Penyelarasan Menuju Fisika-Instrumentasi

Untuk memperkuat identitas instrumentasi, fokus analisis dialihkan dari sekadar fungsionalitas program menjadi **Metrologi, Rantai Pengukuran (*Measurement Chain*), dan Karakteristik Fisik Sensor**:

1. **Perluasan Definisi "Data Integrity" ke Ranah Metrologi:**
   * **Analisis Derau (Noise Modeling):** Mengidentifikasi sumber derau fisik secara spesifik (derau termal resistor, interferensi elektromagnetik/EMI lingkungan, atau *quantization noise* ADC).
   * **Metrik Kinerja Pengukuran:** Evaluasi menggunakan parameter standar instrumentasi seperti **Signal-to-Noise Ratio (SNR)** sebelum dan sesudah filter EMA, serta analisis **akurasi, presisi, dan *phase lag*** (keterlambatan respons sensor terhadap perubahan lingkungan drastis).
2. **Bedah Peran EMA Filter sebagai *Signal Conditioning*:**
   * Memandang EMA sebagai *Low-Pass Filter* (LPF) digital yang memiliki fungsi alih (*transfer function*) tertentu untuk meredam *transient spikes* tanpa merusak tren data fisis yang valid.
3. **Tinjauan Non-Volatile Buffers dari Sudut Pandang Keandalan Perangkat Keras:**
   * **Power Failure & Brownout Physics:** Menganalisis perilaku sistem saat terjadi *voltage drop*, termasuk mengukur waktu transien persis sebelum mikrokontroler mati agar rutinitas *flush* ke memori non-volatif sukses tereksekusi.
   * **Wear Leveling & Material Limits:** Meninjau batasan siklus tulis (*write endurance*) dari memori non-volatif dari perspektif fisika material semikonduktor.

---

## 3. Perangkat Keras dan Lunak untuk Analisis (*Test Bench*)

Untuk mendukung karakterisasi sistem pengukuran dan integritas data secara empiris, dibutuhkan perangkat pengujian berikut:

### A. Perangkat Keras Pengukuran (*Hardware Test Bench*)
* **Digital Storage Oscilloscope (DSO):** Mengambil sampel kurva tegangan (*voltage waveform*) saat terjadi *power failure* atau *brownout* guna mengukur waktu transien dan jeda waktu hingga mikrokontroler padam total.
* **Precision Power Supply / Power Monitor:** Menyediakan suplai daya stabil dan memantau lonjakan arus (*current spike*) saat operasi penulisan (*write cycles*) intensif ke flash/EEPROM.
* **Sensor Acuan (*Ground Truth Instrument*):** Alat ukur standar laboratorium yang telah dikalibrasi sebagai pembanding absolut untuk menghitung parameter kesalahan (*error rate*) dan deviasi data.

### B. Perangkat Lunak & Lingkungan Analisis (*Software Tools*)
* **Python Environment (NumPy, SciPy, Pandas, Matplotlib):** Komputasi *offline* untuk menghitung metrik statistik seperti SNR, *Root Mean Square Error* (RMSE), koefisien korelasi, dan visualisasi respons filter EMA serta *phase lag*.
* **Fault Injection & Automated Stress Test Tools:** Skrip otomatisasi (menggunakan relay atau mikrokontroler sekunder) untuk memutus catu daya secara berkala guna menguji *data recovery rate* dari *non-volatile buffer*.
* **Serial Logger / Terminal Berbasis Script:** Merekam seluruh *debug output* dan aliran data berstempel waktu (*timestamped logs*) selama pengujian ketahanan jangka panjang (*endurance testing*).
