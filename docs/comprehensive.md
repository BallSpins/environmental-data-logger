Ini adalah comprehensive dari seluruh catatan markdown ku:

# Catatan Teknis: Metodologi Penentuan dan Penyetelan Parameter Filter EMA

## 1. Konteks Pengujian & Noise Floor
Sebelum menentukan parameter filter, perangkat keras (misal sensor AHT10) harus dikarakterisasi untuk mengetahui batas galat murni (*Noise Floor*):
*   **Linear Detrending:** Diperlukan untuk memisahkan pergeseran termal lambat (*slow thermal drift* akibat *Joule heating* mikrokontroler) dari osilasi derau murni perangkat keras.
*   **Simpangan Baku Murni ($\sigma$):** Setelah tren linear dibuang, sisa deviasi statistik ($\sigma$) merepresentasikan batas ketidakpastian perangkat keras murni (contoh pada pengujian: $\sigma = 0.0108^\circ\text{C}$).

## 2. Karakteristik & Mekanisme Filter EMA
*   **Sifat Rekursif:** Filter *Exponential Moving Average* (EMA) tidak memerlukan sekumpulan data besar atau banyak sensor; ia hanya membutuhkan nilai pembacaan sensor saat ini ($Y_t$) dan hasil kalkulasi filter pada siklus sebelumnya ($S_{t-1}$).
*   **Persamaan Matematis:**
    $$S_t = \alpha \cdot Y_t + (1 - \alpha) \cdot S_{t-1}$$

## 3. Kalkulasi Persentase Redaman Derau
Redaman derau acak putih (*white noise*) oleh filter EMA dapat diukur secara teoretis menggunakan rasio simpangan baku sebelum dan sesudah difilter:
1. **Faktor Sisa Derau:**
   $$\text{Faktor Sisa} = \sqrt{\frac{\alpha}{2 - \alpha}}$$
2. **Persentase Redaman (Derau Dihancurkan):**
   $$\text{Redaman (\%)} = 100\% - \left( \sqrt{\frac{\alpha}{2 - \alpha}} \times 100\% \right)$$
*Contoh:* Menggunakan $\alpha = 0.1$ akan meredam sekitar **77.06%** derau perangkat keras.

## 4. Konsep *Phase Lag* (Keterlambatan Fase)
*   **Definisi:** *Phase lag* bukanlah keterlambatan pemrosesan CPU, melainkan keterlambatan algoritmik akibat redaman filter yang menahan perubahan nilai demi menjaga kestabilan sinyal.
*   **Trade-off:** Semakin kecil nilai $\alpha$ (semakin banyak derau yang dihancurkan), semakin besar *phase lag* yang terjadi, sehingga sensor menjadi lambat merespons perubahan fisik yang nyata.

## 5. Rumus Optimasi Alpha ($\alpha$) Berdasarkan Batas Toleransi Waktu
Untuk menentukan nilai $\alpha$ secara objektif berdasarkan batas maksimal waktu keterlambatan respons (*maximum allowable delay* / $\tau_{\text{max}}$) dan interval *sampling* ($\Delta t$):
$$\alpha = \frac{\Delta t}{\tau_{\text{max}}}$$
*Contoh:* Jika interval data $\Delta t = 2$ detik dan batas toleransi keterlambatan sistem adalah $\tau_{\text{max}} = 30$ detik, maka nilai $\alpha$ optimal adalah:
$$\alpha = \frac{2}{30} \approx 0.066$$

## 6. Panduan Penentuan Nilai $\alpha$ Secara Bijak (*Engineering Trade-Off*)
Pemilihan nilai $\alpha$ harus disesuaikan dengan dinamika lingkungan dan fungsi sistem:
*   **Sistem / Lingkungan Lambat (Contoh: Suhu Lab/Ruangan):** Cenderung stabil dan tidak berubah ekstrem dalam hitungan detik. Toleran terhadap *phase lag* 20–30 detik. Nilai $\alpha$ di kisaran $0.05$ hingga $0.1$ adalah titik optimal (*sweet spot*) untuk membersihkan grafik dari derau LSB.
*   **Sistem / Lingkungan Cepat / Tinggi Transien (Contoh: Suhu Rem F1):** Berubah drastis dalam fraksi detik. Membutuhkan *phase lag* seminimal mungkin ($\alpha$ mendekati $1.0$), sehingga penggunaan filter agresif akan merusak integritas data puncak.

## 7. Validasi Empiris untuk Laporan Ilmiah
Dalam penulisan naskah riset atau dokumentasi profesional, pemilihan nilai $\alpha$ harus divalidasi secara empiris dengan cara:
1. Mensimulasikan beberapa nilai $\alpha$ yang berbeda (misalnya $0.05$, $0.1$, $0.3$) terhadap dataset nyata di Python.
2. Memplot kurva hasil filter tersebut secara berdampingan (*overlay*).
3. Menyertakan analisis komparatif dalam laporan untuk membuktikan alasan pemilihan parameter secara objektif.

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

# Catatan Eksperimen: Karakterisasi Derau AHT10 (Protokol 1)

## Konteks Pengujian
*   **Target:** Mengukur ketidakpastian perangkat keras (Noise Floor) AHT10 pada kondisi tunak.
*   **Metode:** Sensor diisolasi ke dalam wadah kardus berlapis kain (tas ransel) untuk mengeliminasi konveksi dan fluktuasi termal eksternal.
*   **Durasi Akuisisi:** ~2 jam (Interval sampling 2 detik, komputasi murni di Core 0 ESP32 tanpa interupsi radio Wi-Fi/Bluetooth).

## Analisis Metrologi dan Komputasi
Pengolahan data mentah (raw data) dilakukan menggunakan Python dengan protokol berikut:
1.  **Trimming:** Eliminasi data 30 menit pertama untuk mengompensasi fase lonjakan suhu awal akibat *chamber heating*.
2.  **Detrending Linear:** Sisa data dari durasi pengujian menunjukkan adanya pergerakan termal lambat (*slow thermal drift*) dengan *slope* +0.0043°C/menit akibat akumulasi disipasi panas (*Joule heating*) dari mikrokontroler. Sinyal ini di-detrending untuk mengisolasi tren makroskopis lingkungan dari derau mikroskopis perangkat keras.

## Hasil Karakterisasi (Pasca-Detrending)
*   **Suhu Rata-rata (μ):** 30.7550°C
*   **Laju Drift Lingkungan:** 0.0043°C/menit
*   **Simpangan Baku Murni (σ):** 0.0108°C

## Konklusi Fisis dan Rekayasa
1.  **Batas Resolusi Perangkat:** Nilai $\sigma$ sebesar 0.0108°C mewakili derau listrik murni dan galat kuantisasi pada tingkat LSB (*Least Significant Bit*) dari transduser AHT10.
2.  **Kebutuhan Penkondisian Sinyal:** Tingkat derau ini mendiskualifikasi transmisi data mentah secara langsung. Sinyal mentah wajib dikenai proses penapisan digital (*digital filtering*), seperti Exponential Moving Average (EMA), secara luring di *edge node* sebelum ditransmisikan ke infrastruktur *backend*. Parameter osilasi $\pm$0.0108°C ini akan menjadi landasan untuk penyetelan konstanta penghalus ($\alpha$) pada algoritma filter.

# Dokumentasi Pengujian Protokol 2: Respons Transien dan Evaluasi Phase Lag Filter EMA

## 1. Tujuan Pengujian
Pengujian ini bertujuan untuk memvalidasi efek *thermal lag* (keterlambatan respons fisis) yang dihasilkan oleh penerapan filter *Exponential Moving Average* (EMA) pada sinyal sensor AHT10. Melalui injeksi perubahan suhu mendadak (*step input*), pengujian ini mengevaluasi titik keseimbangan (*trade-off*) yang paling optimal antara kemampuan reduksi derau dan kecepatan respons instrumen.

## 2. Metodologi Akuisisi Data
Data diakuisisi secara kontinu tanpa henti menggunakan mikrokontroler (Edge Node) yang mengirimkan data mentah (*raw data*) langsung ke komputer via antarmuka serial. Prosedur perekaman fisis dibagi menjadi tiga fase:

1.  **Fase Baseline (Menit 0 – 3):**
    Sensor diletakkan di dalam wadah *thinwall* terbuka pada suhu ruangan. Fase ini bertujuan untuk merekam garis dasar (*baseline*) suhu yang stabil tanpa ada akumulasi panas dari mikrokontroler (*Joule heating*).
2.  **Fase Injeksi Transien (Detik ke-180):**
    Sebuah botol berisi air hangat (40°C–50°C) dimasukkan ke dalam wadah, dan wadah langsung ditutup rapat. Tindakan ini menciptakan loncatan suhu (*step input*) yang instan dan terisolasi dari lingkungan luar.
3.  **Fase Ekuilibrium Baru (Menit 3 – 12):**
    Sistem dibiarkan merekam proses rambatan panas di dalam wadah tertutup hingga kurva pembacaan kembali melandai pada titik ekuilibrium suhu yang baru.

## 3. Parameter Komputasi (Post-Compute)
Analisis dilakukan secara *post-compute* menggunakan perangkat lunak Python. Data mentah disimulasikan menggunakan persamaan filter EMA:
$$S_t = \alpha \cdot Y_t + (1 - \alpha) \cdot S_{t-1}$$

Lima parameter konstanta penghalus ($\alpha$) diuji secara simultan untuk membandingkan redaman dan *phase lag*, berdasarkan toleransi batas waktu ($\tau_{\text{max}}$) dengan interval *sampling* ($\Delta t = 2$ detik):
*   $\alpha = 0.05$ ($\tau_{\text{max}} \approx 40$ detik)
*   $\alpha = 0.0666$ ($\tau_{\text{max}} \approx 30$ detik)
*   $\alpha = 0.1$ ($\tau_{\text{max}} \approx 20$ detik)
*   $\alpha = 0.2$ ($\tau_{\text{max}} \approx 10$ detik)
*   $\alpha = 0.4$ ($\tau_{\text{max}} \approx 5$ detik)

## 4. Analisis dan Pembahasan
Berdasarkan visualisasi grafik *overlay* (Keseluruhan dan Zoom Transisi), diperoleh temuan teknis sebagai berikut:

*   **Validitas Baseline & Step Input:**
    Garis data mentah (abu-abu) pada menit 0 hingga 3 menunjukkan kurva yang datar di kisaran 29.9°C, mengonfirmasi tidak adanya *thermal drift*. Injeksi termal pada menit ke-3 terekam sebagai titik infleksi (*inflection point*) yang sangat tajam, membuktikan keberhasilan rekayasa *step input*.
*   **Analisis Filter Lemah (Under-filtered):**
    Pada parameter $\alpha = 0.4$ (kurva merah) dan $\alpha = 0.2$ (kurva oranye), filter memiliki kecepatan respons yang sangat baik (hampir menempel pada data mentah). Namun, galat guncangan LSB (*Least Significant Bit*) dari sensor masih terlihat dengan jelas, sehingga tujuan utama penapisan sinyal tidak tercapai.
*   **Analisis Filter Agresif (Over-filtered):**
    Pada parameter $\alpha = 0.05$ (kurva ungu) dan $\alpha = 0.0666$ (kurva biru), sinyal berhasil diratakan dengan sempurna. Konsekuensinya, *phase lag* yang terjadi sangat masif. Terdapat deviasi horizontal dan vertikal yang signifikan sesaat setelah *step input* terjadi, menunda sistem untuk menyadari perubahan fisis yang aktual.
*   **Titik Keseimbangan Optimal:**
    Parameter $\alpha = 0.1$ (kurva hijau) menunjukkan keseimbangan rasional. Kurva ini meredam osilasi acak dengan persentase teoretis $\approx 77\%$, sekaligus menjaga *phase lag* pada batas toleransi logis ($\approx 20$ detik), sehingga integritas tren kenaikan suhu fisis tetap dapat diobservasi tanpa keterlambatan informasi yang berisiko.

## 5. Kesimpulan Teknis
Pengujian empiris respons transien membuktikan bahwa penentuan nilai $\alpha$ pada filter digital harus disesuaikan dengan *bandwidth* fisis fenomena yang diukur. Untuk instrumen pencatat data lingkungan termal yang tidak bermanuver dalam fraksi detik, nilai **$\alpha = 0.1$** ditetapkan sebagai arsitektur parameter yang paling optimal untuk diimplementasikan secara permanen pada mikrokontroler (*Edge Computing*).

# Optimasi Parameter Filter EMA Berbasis Fungsi Biaya (Cost Function)

## 1. Latar Belakang dan Rasionalisasi Matematis
Penentuan nilai $\alpha$ pada iterasi sebelumnya ($\alpha = 0.1$) masih didasarkan pada metode heuristik dan inspeksi visual terhadap kompromi antara redaman derau (Protokol 1) dan *phase lag* (Protokol 2). Pendekatan tersebut memiliki kelemahan metodologis karena menyisakan ruang bagi bias observasi manusia.

Untuk menetapkan parameter secara absolut, masalah ini direduksi menjadi persoalan optimasi matematis menggunakan Fungsi Biaya (*Cost Function*). Fungsi objektif $J(\alpha)$ dirancang untuk mencari titik minimum global dari akumulasi dua penalti kerugian sistem:
1.  **Penalti Derau ($C_{\text{noise}}$):** Berbanding lurus dengan $\alpha$, dievaluasi menggunakan simpangan baku dari dataset kondisi tunak (Protokol 1).
2.  **Penalti Phase Lag ($C_{\text{lag}}$):** Berbanding terbalik dengan $\alpha$, dievaluasi menggunakan *Root Mean Square Error* (RMSE) dari dataset transien (Protokol 2).

Persamaan Fungsi Biaya yang dinormalisasi:
$$J(\alpha) = W_1 \cdot \hat{C}_{\text{noise}}(\alpha) + W_2 \cdot \hat{C}_{\text{lag}}(\alpha)$$
di mana $W_1 = 0.5$ dan $W_2 = 0.5$ merupakan bobot penyeimbang simetris untuk menjaga prioritas yang setara antara integritas sinyal dan kecepatan respons sistem.

## 2. Hasil Komputasi Numerik
Evaluasi numerik iteratif pada rentang batas $\alpha \in [0.01, 0.99]$ dengan resolusi matriks tinggi membuktikan konvergensi kurva parabola pada titik minimum absolut berikut:
*   **Nilai Optimal Absolut:** $\alpha = 0.1494$
*   **Konstanta Waktu:** $\tau_{\text{max}} = \frac{\Delta t}{\alpha} = \frac{2}{0.1494} \approx 13.38 \text{ detik}$

## 3. Analisis dan Signifikansi Metrologi
Hasil komputasi matematis ini mengoreksi estimasi heuristik sebelumnya dan mengukuhkan argumen rekayasa yang lebih solid:
1.  **Koreksi Titik Buta (Blind Spot):** Secara visual, kurva awal $\alpha = 0.1$ tampak meredam derau dengan baik namun menyisakan *lag* di awal fase transisi, sementara $\alpha = 0.2$ memiliki kecepatan ideal namun terbukti melewatkan derau LSB. Komputasi turunan limit menempatkan titik keseimbangan persis di $\alpha = 0.1494$, membuktikan ketidakmampuan tebakan interval diskret manusia dalam mencari titik optimal presisi.
2.  **Keselarasan Dinamika Fisis:** Konstanta waktu filter $\tau_{\text{max}} \approx 13.4 \text{ detik}$ ini terkalibrasi secara sempurna dengan rentang limit respons intrinsik perangkat keras (8–30 detik). Algoritma filter diatur untuk mengekstrak garis tren terbersih tanpa pernah memproses data lebih lambat dari laju inersia termal sensor itu sendiri.

## 4. Kesimpulan Final
Nilai $\alpha = 0.1494$ ditetapkan sebagai parameter operasional final untuk filter *Exponential Moving Average* pada instrumen *edge node* AHT10. Ketetapan ini tidak lagi bergantung pada interpretasi komparatif visual, melainkan pada pembuktian matematis (minimisasi galat *Cost Function*) yang dapat direplikasi dan dipertanggungjawabkan secara objektif.

# Evaluasi Kuantitatif Phase Lag Filter EMA (Protokol 2 - Teroptimasi)

## 1. Konteks Komputasi
Evaluasi ini merupakan suplemen kuantitatif lanjutan yang mengintegrasikan hasil pencarian titik optimum absolut berbasis Fungsi Biaya (*Cost Function*) kalkulus. Komputasi dilakukan pada rentang waktu kritis pasca-injeksi termal (menit 3.0 hingga 6.0) untuk mengukur deviasi absolut antara sinyal hasil filter (EMA) dengan data suhu mentah (*raw data*).

Metrik yang dihitung meliputi:
*   **Root Mean Square Error (RMSE):** Rata-rata akar kuadrat galat selama jendela transien.
*   **Max Lag Error:** Deviasi absolut terbesar (keterlambatan fase maksimal) antara kurva filter dan data aktual.

## 2. Tabel Hasil Komputasi Numerik

| Parameter ($\alpha$) | Toleransi Lag ($\tau_{\text{max}}$) | RMSE (°C) | Max Lag Error (°C) | Waktu Max Error Terjadi (Menit) |
| :--- | :--- | :--- | :--- | :--- |
| **0.05** | $\approx$ 40 detik | 0.2192 | 0.2686 | 5.61 |
| **0.0666** | $\approx$ 30 detik | 0.1713 | 0.2040 | 5.61 |
| **0.1** | $\approx$ 20 detik | 0.1167 | 0.1585 | 3.39 |
| **0.1494** | **$\approx$ 13.4 detik** | **0.0771** | **0.1333** | **3.39** |
| **0.2** | $\approx$ 10 detik | 0.0558 | 0.1119 | 3.39 |
| **0.4** | $\approx$ 5 detik | 0.0226 | 0.0616 | 3.35 |

## 3. Interpretasi Metrologi

Data numerik di atas memberikan landasan matematis yang objektif — bebas dari bias tebakan manusia — untuk mengevaluasi kinerja filter:

**A. Verifikasi Titik Minimum Absolut ($\alpha = 0.1494$)**
Integrasi pendekatan kalkulus (*Cost Function*) menempatkan parameter optimal di angka $\alpha = 0.1494$. Berdasarkan tabel di atas, nilai ini menghasilkan penurunan RMSE yang signifikan hingga **0.0771°C** dan menekan *Max Lag Error* ke angka **0.1333°C**, jauh lebih baik dibandingkan parameter heuristik awal ($\alpha = 0.1$).

**B. Konsistensi Waktu Respons (Temporal Locking)**
Kolom waktu galat maksimal menunjukkan fenomena struktural yang penting:
*   Pada filter lambat ($\alpha = 0.05$ dan $0.0666$), akumulasi galat terbesar terjadi di penghujung jendela observasi (menit 5.61), membuktikan kegagalan algoritma dalam mengejar laju perubahan fisis suhu.
*   Mulai dari $\alpha = 0.1$ hingga $\alpha = 0.4$, titik galat maksimal secara konsisten **"terkunci" di awal transisi (menit 3.35 – 3.39)**. Ini mengindikasikan bahwa galat tersebut murni karakteristik alami *low-pass filter* yang menahan kejutan sesaat (*shock response*) di detik-detik awal, setelah itu algoritma langsung beradaptasi secara presisi.

**C. Konklusi Optimasi**
Parameter $\alpha = 0.1494$ terbukti menjadi titik keseimbangan absolut. Ia menahan deviasi terburuk di bawah ambang batas 0.15°C, memangkas galat keseluruhan (RMSE), dan mempertahankan keselarasan konstan dengan *time constant* fisis intrinsik perangkat keras tanpa mengorbankan integritas peredaman derau.

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