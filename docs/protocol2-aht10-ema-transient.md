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