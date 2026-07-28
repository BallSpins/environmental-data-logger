### 1. Eksekusi Karakterisasi Derau Dasar (Protokol 1)

Langkah ini bertujuan mencari ketidakpastian pengukuran dan *Noise Floor* dari transduser pada kondisi tunak.

* Tempatkan sensor AHT10 di dalam wadah tertutup yang terisolasi dari fluktuasi angin atau panas ekstrem.


* Konfigurasi mikrokontroler dengan mengisolasi proses akuisisi secara ketat pada **Core 0** untuk meniadakan *jitter* dan memastikan frekuensi *sampling* deterministik.


* Sebagai pengganti W25Q32, evakuasi data langsung menggunakan kabel USB. Cetak *timestamp* dan *raw data* kontinu ke *Serial Monitor* di komputer.
* Masukkan data tersebut ke lingkungan Python untuk menghitung parameter statistik dasar seperti varians dan simpangan baku ($\sigma$).



### 2. Eksekusi Pengujian Respons Transien (Protokol 2)

Langkah ini memvalidasi efek *thermal lag* akibat filter digital.

* Implementasikan persamaan matematis filter *Exponential Moving Average* (EMA) di dalam komputasi Core 0:

$$y_t = \alpha x_t + (1-\alpha)y_{t-1}$$


* Siapkan dua kondisi lingkungan teruji yang berbeda secara mendadak.


* Rekam respons *step input* dari AHT10 saat mengalami transisi suhu terkontrol tersebut.


* Terapkan variasi konstanta penghalus (misalnya $\alpha = 0.1$, $\alpha = 0.5$, $\alpha = 0.9$) dan bandingkan kurva responsnya terhadap data mentah untuk mengukur besar *phase lag*.



### 3. Pembangunan Lingkungan Analisis (*Software Test Bench*)

Gunakan waktu tunggu kedatangan perangkat keras untuk menyelesaikan infrastruktur analisis data di komputer.

* Susun skrip komputasi menggunakan lingkungan Python (termasuk NumPy, SciPy, Pandas, dan Matplotlib).


* Bangun fungsi otomatis untuk memvisualisasikan kurva dan menghitung peningkatan *Signal-to-Noise Ratio* (SNR) serta metrik deviasi *Root Mean Square Error* (RMSE).


## Protokol 1: Karakterisasi Derau Dasar (Kondisi Tunak)

**Durasi Ideal:** 1 hingga 2 jam.

**Justifikasi Metrologi:**

* **Keseimbangan Termal:** Saat Anda memasukkan sensor ke dalam wadah tertutup, volume udara di dalam wadah membutuhkan waktu untuk mencapai ekuilibrium absolut. Jika Anda menghitung varians saat suhu di dalam wadah masih bergerak menuju kestabilan, metrik *Noise Floor* Anda akan cacat.
* **Kecukupan Sampel Statistik:** Untuk mencegah degradasi akurasi akibat kenaikan suhu internal cip (*self-heating*), interval pengukuran yang direkomendasikan adalah setiap 2 detik. Dengan interval 2 detik, durasi 1 jam akan menghasilkan 1800 titik data. Jumlah populasi sampel ini sudah sangat valid secara statistik untuk membentuk kurva distribusi Gaussian dan mencari simpangan baku ($\sigma$).



## Protokol 2: Pengujian Respons Transien (*Phase Lag*)

**Durasi Ideal:** 10 hingga 15 menit per siklus perubahan (tidak perlu berjam-jam).

**Justifikasi Metrologi:**

* **Fokus Akuisisi:** Parameter yang dicari adalah kelambatan respons (*thermal lag*) filter EMA. Momen kritis dalam pengujian ini hanya terjadi pada detik-detik pertama saat lingkungan fisis berubah drastis (*step input*).
* **Karakteristik Perangkat Keras:** Waktu respons perangkat untuk mencapai titik 63% dari nilai akhir lingkungan berada pada rentang 8 hingga 30 detik. Setelah grafik sinyal kembali melandai dan mencapai kondisi stabil pada suhu yang baru (umumnya memakan waktu kurang dari 10 menit), siklus pengujian tersebut secara fisis sudah selesai. Merekam lebih lama dari itu hanya membuang kapasitas memori tanpa memberikan informasi baru mengenai *phase lag*.

