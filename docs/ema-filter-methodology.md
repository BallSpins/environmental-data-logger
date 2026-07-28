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