# Dokumentasi Pengujian Protokol 2: Respons Transien dan Evaluasi Phase Lag Filter EMA

## 1. Tujuan Pengujian
Pengujian ini bertujuan untuk memvalidasi efek *thermal lag* (keterlambatan respons fisis) yang dihasilkan oleh penerapan filter *Exponential Moving Average* (EMA) pada sinyal sensor AHT10, untuk kanal suhu (T) maupun kelembapan relatif (RH). Melalui injeksi perubahan suhu mendadak (*step input*), pengujian ini mengevaluasi titik keseimbangan (*trade-off*) awal secara heuristik/visual antara kemampuan reduksi derau dan kecepatan respons instrumen, sebagai dasar sebelum optimasi matematis formal.

## 2. Metodologi Akuisisi Data
Data diakuisisi secara kontinu menggunakan mikrokontroler (Edge Node) yang mengirimkan data mentah langsung ke komputer via antarmuka serial, dengan interval sampling terkunci *hardware timer* (nominal 2 detik; rata-rata terukur 2.051 detik, jitter <0.05 ms). Prosedur perekaman fisis dibagi menjadi tiga fase:

1.  **Fase Baseline (Menit 0 – 3):**
    Sensor diletakkan di dalam wadah *thinwall* terbuka pada suhu ruangan, merekam garis dasar (*baseline*) suhu dan RH stabil tanpa akumulasi panas dari mikrokontroler (*Joule heating*).
2.  **Fase Injeksi Transien (Detik ke-180):**
    Sebuah botol berisi air hangat (40°C–50°C) dimasukkan ke dalam wadah, dan wadah langsung ditutup rapat, menciptakan loncatan suhu (*step input*) yang instan dan terisolasi dari lingkungan luar. Karena RH bergantung pada suhu, loncatan ini juga menciptakan loncatan RH secara tidak langsung.
3.  **Fase Ekuilibrium Baru (Menit 3 – 12):**
    Sistem merekam proses rambatan panas dan kelembapan di dalam wadah tertutup hingga kurva pembacaan T dan RH kembali melandai pada titik ekuilibrium yang baru.

## 3. Parameter Komputasi (Post-Compute)
Analisis dilakukan secara *post-compute* menggunakan Python, diterapkan identik pada kedua kanal. Data mentah diproses menggunakan persamaan filter EMA:
$$S_t = \alpha \cdot Y_t + (1 - \alpha) \cdot S_{t-1}$$

Lima parameter konstanta penghalus ($\alpha$) diuji secara simultan. Konstanta waktu ($\tau_{\text{max}}$) dihitung menggunakan rumus eksak $\tau = -\Delta t/\ln(1-\alpha)$ dengan interval sampling terukur $\Delta t = 2.051$ detik:

*   $\alpha = 0.05$ ($\tau_{\text{max}} \approx 40.0$ detik)
*   $\alpha = 0.0666$ ($\tau_{\text{max}} \approx 29.8$ detik)
*   $\alpha = 0.1$ ($\tau_{\text{max}} \approx 19.5$ detik)
*   $\alpha = 0.2$ ($\tau_{\text{max}} \approx 9.2$ detik)
*   $\alpha = 0.4$ ($\tau_{\text{max}} \approx 4.0$ detik)

## 4. Analisis dan Pembahasan

### 4.1 Kanal Suhu (T)
*   **Validitas Baseline & Step Input:**
    Data mentah pada menit 0–3 berada di kisaran 29.86–29.93°C dengan slope regresi lemah dan tidak signifikan ($R^2 = 0.49$), mengonfirmasi tidak adanya *thermal drift* sistematis pada fase baseline. Injeksi termal pada menit ke-3 terekam sebagai titik infleksi tajam, membuktikan keberhasilan rekayasa *step input*.
*   **Analisis Filter Lemah (*Under-filtered*):**
    Pada $\alpha = 0.4$ dan $\alpha = 0.2$, filter memiliki kecepatan respons sangat baik, tetapi galat guncangan LSB masih terlihat jelas sehingga tujuan penapisan sinyal tidak tercapai.
*   **Analisis Filter Agresif (*Over-filtered*):**
    Pada $\alpha = 0.05$ dan $\alpha = 0.0666$, sinyal diratakan dengan baik, namun *phase lag* yang terjadi cukup besar, menunda sistem menyadari perubahan fisis aktual.
*   **Titik Keseimbangan Heuristik:**
    Parameter $\alpha = 0.1$ menunjukkan keseimbangan yang secara visual rasional: reduksi standar deviasi teoretis $\approx 77.1\%$ (rasio varians $\alpha/(2-\alpha) = 0.0526$), sekaligus menjaga *phase lag* pada $\tau_{\text{max}} \approx 19.5$ detik. **Penentuan ini masih bersifat heuristik/visual.**

### 4.2 Kanal Kelembapan (RH)
*   **Validitas Baseline & Step Input:**
    Data mentah RH pada menit 0–3 berada di kisaran 60.48–60.89 %RH. Slope regresi pada rentang ini ($R^2 = 0.71$) sedikit lebih kentara dibanding kanal suhu ($R^2=0.49$), namun rentang absolutnya tetap kecil (<0.5 %RH selama 3 menit), sehingga fase baseline tetap dapat dianggap cukup stabil untuk titik acuan *step input*. Injeksi termal pada menit ke-3 memicu **kenaikan** RH (bukan penurunan), dari ≈60.5 %RH menuju puncak ≈86 %RH — konsisten dengan uap air dari botol air hangat yang turut menambah kelembapan absolut di dalam wadah tertutup, di luar efek murni suhu.
*   **Skala Galat Jauh Lebih Besar dari Kanal Suhu:**
    Karena rentang perubahan RH selama transien (≈26 %RH) jauh lebih besar dibanding rentang perubahan suhu (≈1.5°C), RMSE dan *Max Lag Error* pada kanal RH juga jauh lebih besar secara nominal (orde 1–5 %RH, dibanding orde 0.02–0.3°C pada kanal suhu). Ini bukan indikasi kanal RH "lebih berisik", melainkan konsekuensi dari skala fisis kejadian yang direkam.
*   **Titik Keseimbangan Heuristik:**
    Berbeda dengan kanal suhu, filter kelembapan menunjukkan pola optimal pada nilai $\alpha$ yang lebih besar (lihat `protocol2-calculus-optimization.md` untuk nilai presisi $\alpha_{RH}=0.2260$) — mengindikasikan bahwa dinamika transien RH pada instrumen ini secara intrinsik lebih cepat dibanding suhu, sehingga menoleransi *smoothing* yang lebih agresif tanpa mengorbankan *lag* secara proporsional.

## 5. Kesimpulan Teknis (Historis — Digantikan)
Pada tahap ini, $\alpha = 0.1$ ditetapkan sebagai kandidat terbaik untuk kanal suhu berdasarkan inspeksi visual dan reduksi noise teoretis. **Nilai ini kemudian dikoreksi menjadi $\alpha_T = 0.1494$ dan dilengkapi dengan $\alpha_{RH} = 0.2260$ untuk kanal kelembapan, melalui optimasi Cost Function kalkulus (`protocol2-calculus-optimization.md`), yang menghilangkan bias observasi manusia dari proses penentuan parameter dan memisahkan penyetelan filter per kanal.**