# Catatan Eksperimen: Karakterisasi Derau AHT10 (Protokol 1)

## Konteks Pengujian
*   **Target:** Mengukur ketidakpastian perangkat keras (*Noise Floor*) AHT10 pada kondisi tunak, untuk kanal suhu (T) dan kelembapan relatif (RH).
*   **Metode:** Sensor diisolasi ke dalam wadah kardus berlapis kain (tas ransel) untuk mengeliminasi konveksi dan fluktuasi termal eksternal.
*   **Durasi Akuisisi:** ~2 jam. Interval sampling nominal 2 detik; terukur rata-rata **2.051 detik** dengan simpangan baku jitter hanya **0.02–0.05 ms**. Jitter yang nyaris nol ini memvalidasi klaim akuisisi deterministik berbasis *hardware timer* di Core 0 ESP32, tanpa interupsi radio Wi-Fi/Bluetooth.
*   **Justifikasi Interval 2 Detik:** Pemilihan interval ini bukan nilai arbitrer, melainkan mengikuti rekomendasi eksplisit *datasheet* AHT10 (Bagian "Temperature Effects"): untuk menjaga kenaikan suhu internal sensor akibat swa-pemanasan (*self-heating*) dari siklus pengukuran berulang tetap di bawah 0.1°C, produsen merekomendasikan siklus pengukuran tidak lebih cepat dari sekali setiap 2 detik. Rekomendasi ini relevan untuk kedua kanal, karena pengukuran RH pada AHT10 turut bergantung pada kondisi termal internal sensor.
*   **Cakupan Kanal:** Karakterisasi *noise floor* pada dokumen ini mencakup **kedua kanal** — suhu (`temperature_c`) dan kelembapan relatif (`humidity_rh`) — melalui pipeline *trimming* dan *detrending* yang identik.

## Analisis Metrologi dan Komputasi
Pengolahan data mentah dilakukan menggunakan Python dengan protokol berikut, diterapkan secara independen pada kanal T dan RH:

1.  **Trimming:** Eliminasi data 30 menit pertama akuisisi.
    Pemeriksaan langsung terhadap data mentah menunjukkan bahwa 30 menit pertama ini **bukan** merupakan lonjakan suhu naik ("*chamber heating*"), melainkan sebuah **transien pendinginan**: suhu turun dari ≈30.99°C menuju titik minimum ≈30.56°C pada menit ke-26–28, kemungkinan besar akibat disipasi sisa panas dari proses penanganan/instalasi sensor sesaat sebelum pengujian dimulai. Segmen ini dibuang bukan untuk menghindari "pemanasan *chamber*", melainkan untuk memastikan karakterisasi *noise floor* hanya dilakukan setelah sistem mencapai rezim termal kuasi-linear (*near-steady-state*).
2.  **Detrending Linear:** Sisa data pasca-30-menit menunjukkan pergerakan lambat pada kedua kanal:
    *   **Suhu:** *slope* **+0.0043°C/menit** ($R^2 = 0.991$), kemungkinan akibat kombinasi akumulasi disipasi panas (*Joule heating*) dari mikrokontroler dan pemanasan ambien ruangan. Fit kuadratik pembanding hanya meningkatkan $R^2$ ke 0.9915 — peningkatan tidak signifikan — sehingga **detrending linear terbukti memadai secara statistik**.
    *   **Kelembapan:** *slope* **−0.010018 %RH/menit** ($R^2 = 0.980$), arah berlawanan dengan suhu — konsisten secara fisis, karena RH berbanding terbalik dengan suhu pada kandungan uap air konstan (semakin hangat wadah, semakin rendah RH relatifnya).
    Kedua sinyal di-detrending untuk mengisolasi tren makroskopis lingkungan dari derau mikroskopis perangkat keras.

## Hasil Karakterisasi (Pasca-Detrending)
*   **Rata-rata:**
    *   Suhu ($\mu_{T}$): 30.7550°C
    *   Kelembapan ($\mu_{RH}$): 53.2896 %RH
*   **Laju Drift Lingkungan:**
    *   +0.0043°C/menit ($R^2 = 0.991$)
    *   −0.010018 %RH/menit ($R^2 = 0.980$)
*   **Simpangan Baku Murni:**
    *   ($\sigma_{T}$): 0.0108°C
    *   ($\sigma_{RH}$): 0.0370 %RH

## Konklusi Fisis dan Rekayasa
1.  **Batas Resolusi Perangkat — Suhu:** Nilai $\sigma_{T} = 0.0108^\circ\text{C}$ dibandingkan terhadap resolusi kuantisasi digital teoretis dari arsitektur ADC 20-bit AHT10:
    $$1 \text{ LSB}_T = \frac{200}{2^{20}} \approx 0.0001907^\circ\text{C}$$
    Rasio $\sigma_T$ terhadap 1 LSB adalah **56.8×** ($0.0108 / 0.0001907 \approx 56.8$).
2.  **Batas Resolusi Perangkat — Kelembapan:** Nilai $\sigma_{RH} = 0.0370\,\%\text{RH}$ dibandingkan terhadap resolusi kuantisasi digital teoretis, memakai formula konversi RH datasheet ($RH[\%] = (S_{RH}/2^{20}) \times 100$):
    $$1 \text{ LSB}_{RH} = \frac{100}{2^{20}} \approx 0.0000954\,\%\text{RH}$$
    Rasio $\sigma_{RH}$ terhadap 1 LSB adalah **388×** ($0.0370 / 0.0000954 \approx 388$) — bahkan lebih dominan dibanding kanal suhu.
    Kedua rasio ini membuktikan bahwa derau terukur pada kedua kanal jauh melampaui batas kuantisasi digital. Temuan ini secara kuat mengindikasikan (bukan sekadar diasumsikan) bahwa batas bawah akurasi sensor pada kondisi tunak didominasi oleh derau termodinamika/analog pada *analog front-end*, bukan galat kuantisasi digital — untuk kedua besaran fisis yang diukur.
3.  **Kebutuhan Pengondisian Sinyal:** Tingkat derau pada kedua kanal mendiskualifikasi transmisi data mentah secara langsung. Sinyal mentah wajib dikenai proses penapisan digital (*digital filtering*), seperti *Exponential Moving Average* (EMA), secara luring di *edge node* sebelum ditransmisikan ke infrastruktur *backend*. Karena karakteristik derau dan dinamika transien T dan RH berbeda (lihat `protocol2-calculus-optimization.md`), **konstanta penghalus dioptimasi secara terpisah untuk tiap kanal** ($\alpha_T$ dan $\alpha_{RH}$), bukan menggunakan satu nilai $\alpha$ tunggal untuk keduanya.

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

# Optimasi Parameter Filter EMA Berbasis Fungsi Biaya (Cost Function)

## 1. Latar Belakang dan Rasionalisasi Matematis
Penentuan nilai $\alpha$ pada iterasi sebelumnya ($\alpha = 0.1$, lihat `protocol2-aht10-ema-transient.md`) masih didasarkan pada metode heuristik dan inspeksi visual terhadap kompromi antara redaman derau (Protokol 1) dan *phase lag* (Protokol 2), dan hanya dievaluasi untuk kanal suhu. Pendekatan tersebut memiliki kelemahan metodologis karena menyisakan ruang bagi bias observasi manusia, dan tidak memperhitungkan bahwa kanal RH punya karakteristik derau dan dinamika transien yang berbeda dari kanal suhu.

Untuk mengurangi ketergantungan pada inspeksi visual, masalah ini direduksi menjadi persoalan optimasi matematis menggunakan Fungsi Biaya (*Cost Function*), yang dijalankan secara independen untuk tiap kanal. Fungsi objektif $J(\alpha)$ dirancang untuk mencari titik minimum dari akumulasi dua penalti kerugian sistem:
1.  **Penalti Derau ($C_{\text{noise}}$):** Berbanding lurus dengan $\alpha$, dievaluasi menggunakan simpangan baku dari dataset kondisi tunak (Protokol 1), per kanal.
2.  **Penalti Phase Lag ($C_{\text{lag}}$):** Berbanding terbalik dengan $\alpha$, dievaluasi menggunakan *Root Mean Square Error* (RMSE) dari dataset transien (Protokol 2), per kanal.

Persamaan Fungsi Biaya yang dinormalisasi (identik untuk kedua kanal):
$$J(\alpha) = W_1 \cdot \hat{C}_{\text{noise}}(\alpha) + W_2 \cdot \hat{C}_{\text{lag}}(\alpha)$$
di mana $W_1 = 0.5$ dan $W_2 = 0.5$ merupakan bobot penyeimbang simetris yang **dipilih sebagai asumsi desain studi ini** untuk menjaga prioritas setara antara integritas sinyal dan kecepatan respons sistem. Bobot yang berbeda (misalnya memprioritaskan kecepatan respons untuk aplikasi yang butuh deteksi perubahan cepat) akan menggeser titik optimal ke nilai $\alpha$ yang berbeda.

## 2. Hasil Komputasi Numerik

### 2.1 Kanal Suhu (T)
Evaluasi numerik iteratif pada rentang $\alpha \in [0.01, 0.99]$ (500 titik) menghasilkan kurva $J(\alpha)$ berbentuk parabola dengan titik minimum pada:
*   **Nilai Optimal (relatif terhadap $W_1=W_2=0.5$):** $\alpha_T = 0.1494$
*   **Konstanta Waktu ($\Delta t$ terukur = 2.051 detik):**
    $$\tau_T = \frac{-\Delta t}{\ln(1-\alpha_T)} = \frac{-2.051}{\ln(0.8506)} \approx 12.67 \text{ detik}$$
*   **Reduksi standar deviasi teoretis:** $\approx 71.6\%$ ($\alpha_T/(2-\alpha_T) = 0.0808$)

### 2.2 Kanal Kelembapan (RH)
Dengan metodologi identik, evaluasi untuk kanal RH menghasilkan:
*   **Nilai Optimal (relatif terhadap $W_1=W_2=0.5$):** $\alpha_{RH} = 0.2260$
*   **Konstanta Waktu (rumus eksak, $\Delta t$ terukur = 2.051 detik):**
    $$\tau_{RH} = \frac{-\Delta t}{\ln(1-\alpha_{RH})} = \frac{-2.051}{\ln(0.7740)} \approx 8.00 \text{ detik}$$
*   **Reduksi standar deviasi teoretis:** $\approx 64.3\%$ ($\alpha_{RH}/(2-\alpha_{RH}) = 0.1274$)

Nilai $\alpha_{RH}$ jauh lebih besar dari $\alpha_T$ — Fungsi Biaya RH secara numerik "memilih" filter yang jauh lebih ringan/responsif dibanding filter suhu, karena dinamika transien kelembapan pada pengujian ini lebih cepat relatif terhadap derau tunaknya dibanding suhu.

## 3. Analisis dan Signifikansi Metrologi
1.  **Koreksi Titik Buta (*Blind Spot*):** Secara visual, kurva awal $\alpha = 0.1$ (dipakai untuk suhu) tampak meredam derau dengan baik namun menyisakan *lag* di awal fase transisi, sementara $\alpha = 0.2$ memiliki kecepatan ideal namun terbukti melewatkan derau LSB. Komputasi numerik menempatkan titik keseimbangan (untuk bobot simetris yang dipilih) persis di $\alpha_T = 0.1494$ — menunjukkan bahwa tebakan interval diskret manusia ($\alpha=0.1$ atau $0.2$) tidak sepresisi solusi numerik kontinu. Pola serupa berlaku untuk RH, di mana nilai heuristik lama ($\alpha=0.1$, dirancang untuk suhu) ternyata bukan titik optimal untuk kelembapan.
2.  **Keselarasan Dinamika Fisis — Suhu:** Konstanta waktu filter $\tau_T \approx 12.7$ detik berada dalam rentang limit respons intrinsik perangkat keras yang disebutkan datasheet AHT10 **untuk kanal suhu** — *response time* $t_{63\%}$ = **5–30 detik** (Table 3, "Temperature", kondisi *min–max*). Catatan: angka "8 detik" yang sempat dipakai pada draf sebelumnya untuk klaim ini keliru — itu adalah *typical response time* kanal **kelembapan (RH)** (Table 1), bukan suhu.
3.  **Keselarasan Dinamika Fisis — Kelembapan:** Menariknya, konstanta waktu filter RH hasil optimasi, $\tau_{RH} \approx 8.0$ detik, justru **hampir persis sama** dengan *typical response time* $t_{63\%}$ = 8 detik yang dispesifikasikan datasheet AHT10 untuk kanal RH (Table 1, "Relative humidity", kondisi *typical*). Ini adalah validasi silang yang kuat: optimasi Cost Function murni berbasis data empiris (independen dari datasheet) konvergen ke konstanta waktu yang secara independen juga tercatat sebagai spesifikasi fisis bawaan sensor.
4.  **Konsekuensi Implementasi:** Karena $\alpha_T \neq \alpha_{RH}$, implementasi akhir pada *edge node* memerlukan **dua filter EMA independen** — satu per kanal — bukan satu filter tunggal yang dipakai untuk kedua besaran.

## 4. Kesimpulan
Nilai $\alpha_T = 0.1494$ (suhu) dan $\alpha_{RH} = 0.2260$ (kelembapan) ditetapkan sebagai parameter operasional final untuk filter *Exponential Moving Average* pada instrumen *edge node* AHT10, **di bawah asumsi bobot penalti simetris ($W_1=W_2=0.5$) yang didefinisikan pada studi ini**. Ketetapan ini tidak lagi bergantung pada interpretasi komparatif visual, melainkan pada minimisasi numerik Fungsi Biaya yang dapat direplikasi dan dipertanggungjawabkan secara objektif per kanal — dengan catatan bahwa perubahan prioritas aplikasi (bobot $W_1, W_2$ berbeda) akan menghasilkan titik optimal $\alpha$ yang berbeda pula.

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