# Eksperimen 1: JSON vs Binary Benchmark

Eksperimen ini dirancang untuk membandingkan performa antara format paket biner 48-byte teroptimasi dengan payload standar JSON.

## Metodologi
1. **Isolated Benchmark**: Melakukan serialisasi dan deserialisasi struktur data sensor sebanyak $N$ kali (sesuai flag `--count` dan `--duration`) untuk mendapatkan rata-rata durasi dalam nanodetik per operasi (`ns/op`).
2. **Bandwidth & Compression Analysis**: Menghitung ukuran byte mentah payload dan ukuran setelah dikompresi menggunakan `Gzip` untuk menyimulasikan perbandingan transfer data di dunia nyata.
3. **Memory Metrics**: Melacak jumlah memori heap (`AllocatedMB`) yang dialokasikan selama operasi berlangsung menggunakan fungsi internal `runtime.ReadMemStats`.

## Parameter Konfigurasi
* `--count` (default: 5): Jumlah perulangan pengujian untuk menjamin stabilitas statistik (*repeatability*).
* `--duration` (default: 3s): Batas waktu run per percobaan.
* `--output` (default: "results"): Direktori ekspor laporan.
* `--pprof` (default: false): Ekspor profil CPU dan Heap.

## Cara Menjalankan
```bash
cd backend/
go run cmd/benchmark-json-binary/main.go --count=5 --duration=3s
```

## Interpretasi Metrik
* **Payload Size (B)**: Ukuran biner jauh lebih kecil dibanding JSON karena tidak menyertakan struktur string kunci (keys) dan format representasi teks. Berdampak langsung pada kuota pengiriman data AWS IoT Core.
* **Serialization/Deserialization Time**: Penggunaan `ParserService` biner berkecepatan $O(1)$ tanpa refleksi menghemat siklus CPU secara masif dibanding encoder/decoder JSON bawaan Go.
* **Heap Allocated**: Solusi biner menerapkan optimasi nol alokasi heap (*zero-allocation*), sehingga meredam beban siklus pembersihan memori (*Garbage Collection*) di backend.
