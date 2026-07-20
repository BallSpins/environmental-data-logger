# Eksperimen 4: Redis Buffer vs Direct MySQL

Eksperimen ini menyimulasikan kejadian letupan beban (*burst spike*) untuk membuktikan efektivitas penggunaan Redis FIFO List sebagai penyangga kestabilan (*shock absorber*) untuk meredam lonjakan penulisan langsung ke MySQL.

## Metodologi
1. **Direct Ingestion Scenario**: Mengirimkan $N$ letupan letupan pesan secara konkuren langsung ke MySQL lokal (MQTT langsung memicu single INSERT per data).
2. **Buffered Ingestion Scenario**: Mengirimkan $N$ letupan pesan ke dalam Redis antrean (`queue:sensor_ingestion`) menggunakan `LPUSH`, kemudian dikuras secara bertahap oleh proses worker menggunakan `RPopCount` dengan batch size 100 untuk dimasukkan ke MySQL.
3. Mencatat durasi pemrosesan penuh (`sec`) dan akumulasi antrean maksimum yang terjadi pada Redis (`max queue accumulation`).

## Parameter Konfigurasi
* `--count` (default: 5): Jumlah repetisi pengujian.
* `--burst-rate` (default: 500): Jumlah pesan konkuren yang dikirim secara serentak.
* `--output` (default: "results"): Direktori ekspor laporan.

## Cara Menjalankan
```bash
cd backend/
go run cmd/benchmark-redis/main.go --count=5 --burst-rate=500
```

## Interpretasi Metrik
* **Processing Duration**: Skenario buffer Redis akan jauh lebih cepat merespons pengirim data IoT dibanding menulis langsung ke MySQL yang lambat akibat pembatasan koneksi database dan performa disk SSD.
* **Max Queue Accumulation**: Redis menampung letupan lonjakan pesan dengan sangat aman di memori, bertindak sebagai penampung sementara (*surge tank*), memastikan backend tidak mengalami kegagalan sistem (*crash/Out-Of-Memory*) atau kehilangan data akibat MySQL yang menolak koneksi tambahan (*Too many connections*).
