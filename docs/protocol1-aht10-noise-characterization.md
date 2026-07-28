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