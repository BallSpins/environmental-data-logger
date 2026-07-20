# Eksperimen 5: End-to-End Scalability Stress-Testing

Alat simulasi *stress testing* tingkat lanjut untuk menguji keandalan sistem dari ujung ke ujung (*end-to-end*) dengan mensimulasikan beban ribuan perangkat virtual IoT secara real-time.

## Metodologi
* Membangun **Worker Pool** dinamis dengan **Traffic Scheduler** yang mengatur emulasi publish secara presisi tanpa pemborosan overhead scheduling Go runtime.
* Mendukung 3 model pola trafik telemetri (`--pattern`):
  1. **`constant`**: Pengiriman data secara teratur dan konstan dari setiap perangkat virtual.
  2. **`random`**: Pemilihan perangkat secara acak dengan jitter waktu tambahan untuk meniru ketidakpastian jaringan.
  3. **`burst`**: Pengiriman letupan pesan berskala besar secara berkala diselingi periode hening.
* Melacak distribusi latensi pengiriman (`P50`, `P90`, `P95`, `P99`), akumulasi panjang antrean Redis per detik, laju throughput (pesan/detik), laju bandwidth jaringan (KB/s), dan pemakaian memori heap Go.

## Parameter Konfigurasi
* `--nodes` (default: 100): Jumlah instansi perangkat virtual IoT yang diemulasikan.
* `--duration` (default: 10s): Lama waktu eksekusi stress test berlangsung.
* `--payload` (binary/json): Format payload yang dikirimkan.
* `--publish-rate` (default: 0.05): Frekuensi kirim per detik per perangkat (0.05 = 1 pesan tiap 20 detik).
* `--pattern` (constant/random/burst): Pola sebaran trafik data.
* `--pprof`: Ekspor pprof profile jika bernilai true.

## Cara Menjalankan
```bash
cd backend/
go run cmd/stress-shadow-device/main.go --nodes=1000 --duration=30s --payload=binary --pattern=random
```

## Interpretasi Metrik
* **Latency Percentiles (P50 vs P99)**: Jika latensi rata-rata rendah namun P99 (tail latency) sangat tinggi, ini menunjukkan adanya hambatan internal (*latency spikes*) periodik akibat Garbage Collector atau lock contention database.
* **Periodic Queue Length Tracking**: Memantau grafik pertumbuhan antrean Redis per detik. Jika grafik terus mendaki naik, ini adalah tanda bahwa kapasitas pemrosesan Ingestion Worker di database hilir lebih lambat dibanding laju data hulu yang masuk (*backpressure bottleneck*).
* **Bandwidth & CPU**: Membantu memproyeksikan kapasitas jaringan yang harus disiapkan di AWS EC2 / ECS untuk menangani ribuan perangkat IoT aktif.
