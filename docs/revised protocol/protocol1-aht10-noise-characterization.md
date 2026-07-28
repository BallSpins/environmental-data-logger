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