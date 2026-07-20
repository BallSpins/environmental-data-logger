Target utama sekarang adalah mengambil data metrik yang kedepannya akan dipergunakan untuk menghitung perkiraan AWS cost

| Variabel (X)                           | Metrik yang diukur (Y)                                                     |
| -------------------------------------- | -------------------------------------------------------------------------- |
| JSON → Binary                          | Ukuran payload, waktu serialisasi/deserialisasi, penggunaan CPU, bandwidth, latency, throughput |
| Chunk 1 → 10 → 30 → 60                 | Jumlah MQTT publish, biaya AWS IoT Core, latency end-to-end                |
| Single INSERT → Batch INSERT           | Inserts/s, penggunaan CPU database, waktu commit                           |
| Tanpa Redis → Dengan Redis             | Throughput saat burst, ketahanan terhadap lonjakan beban                   |
