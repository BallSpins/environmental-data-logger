# Eksperimen 2: Chunk Aggregation Benchmark

Eksperimen ini menganalisis efisiensi transmisi data dari node IoT ke backend melalui optimasi akumulasi atau pengelompokan (*aggregation*) beberapa rekaman data sensor pada sisi perangkat keras sebelum dipublikasikan.

## Metodologi
Membandingkan laju throughput transmisi biner dengan memvariasikan ukuran paket log sensor yang terkandung di dalamnya:
* **Size 1**: 1 rekaman sensor per paket (ukuran 12 byte)
* **Size 10**: 10 rekaman sensor per paket (ukuran 48 byte - *Baseline*)
* **Size 30**: 30 rekaman sensor per paket (ukuran 128 byte)
* **Size 60**: 60 rekaman sensor per paket (ukuran 248 byte)

Sistem akan mengukur laju operasi per detik (`ops/sec`) yang mampu dicapai oleh masing-masing ukuran chunk.

## Parameter Konfigurasi
* `--count` (default: 5): Jumlah repetisi pengujian.
* `--duration` (default: 3s): Batas waktu run per percobaan.
* `--output` (default: "results"): Direktori ekspor laporan.

## Cara Menjalankan
```bash
cd backend/
go run cmd/benchmark-chunk/main.go --count=5 --duration=3s
```

## Interpretasi Metrik & Estimasi AWS Cost
* **AWS IoT Core Message Count**: AWS IoT Core membebankan biaya per jutaan pesan yang masuk. Batasan minimum satu pesan di AWS IoT Core adalah 5KB.
* **Perhitungan Kasus**:
  * Jika menggunakan **Chunk 1** (mengirim setiap 2 detik): Perangkat mengirimkan 30 pesan per menit.
  * Jika menggunakan **Chunk 10** (mengirim setiap 20 detik): Perangkat mengirimkan 3 pesan per menit (menghemat **90% biaya AWS IoT Core**).
  * Semakin tinggi ukuran chunk, semakin rendah total message billing AWS IoT Core, namun latensi ketersediaan data secara real-time di server akan bertambah (*latency-cost trade-off*).
