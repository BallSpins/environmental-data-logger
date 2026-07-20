# Eksperimen 3: Single INSERT vs Batch INSERT

Eksperimen ini menguji tingkat efisiensi penulisan data ke dalam MySQL dengan membandingkan operasi satu per satu (*Single INSERT*) melawan penulisan berkelompok (*Batch INSERT*) dengan ukuran batch yang bervariasi.

## Metodologi
1. Membuat dataset teratur sebanyak 1.000 log sensor secara deterministik.
2. Melakukan penulisan ke MySQL lokal dengan ukuran batch: **1 (Single), 10, 25, 50, 100, 250, dan 500**.
3. Melacak laju penulisan baris per detik (`inserts/sec`) dan overhead alokasi memori heap (`AllocatedMB`) untuk setiap ukuran batch.

## Parameter Konfigurasi
* `--count` (default: 5): Jumlah repetisi pengujian (*repeatability*).
* `--output` (default: "results"): Direktori ekspor laporan.
* `--mysql` (opsional): Custom MySQL DSN.

## Cara Menjalankan
```bash
cd backend/
go run cmd/benchmark-db/main.go --count=5
```

## Interpretasi Metrik
* **Inserts/sec (Throughput)**: Penyisipan data per baris (*Single*) memaksa database melakukan siklus commit transaksi disk I/O, negosiasi lock, dan transfer jaringan bolak-balik untuk setiap baris, membatasi performa secara ekstrem.
* **Batching Gains**: Mengelompokkan data ke dalam batch (seperti batch size 100 atau 250) mengonsolidasikan transaksi ke dalam satu pernyataan SQL tunggal. Ini memaksimalkan efisiensi commit disk MySQL dan meminimalkan latensi jaringan.
* **Heap Allocated**: Memastikan bahwa meskipun ukuran batch meningkat hingga 500, alokasi heap dari `SensorRepository.BatchInsert` tetap terkendali karena dioptimasi menggunakan `sync.Pool`.
