size data struct dengan pendekatan biner: 4 byte
```c++
struct Data {
  int16_t humi; // 2 byte
  int16_t temp; // 2 byte
};
```

size data struct dengan pendekatan json: 25 byte (no spaces)
```json
{ // 1 byte
  "humi": 3083, // 12 byte
  "temp": 2450 // 11 byte
} // 1 byte
```

size data struct dengan pendekatan json: 19 byte (no spaces & shorten key)
```json
{ // 1 byte
  "h": 3083, // 9 byte
  "t": 2450 // 8 byte
} // 1 byte
```

dari data diatas, dapat dilihat bahwa pendekatan biner mampu menghemat 15 - 21 byte tergantung pendekatan json mana yang dipilih. data struct sendiri akan digunakan oleh chunk struct, yang dimana pada 1 chunk akan menampung 10 data. jadi, pada 10 data saja, pendekatan biner mampu menghemat 150 - 210 byte.

size chunk struct dengan pendekatan biner:  48 byte
```c++
struct Chunk {
  uint32_t timestamp; // 4 byte
  uint8_t nodeID; // 1 byte
  uint8_t padding[3]; // 3 byte
  Data data[10]; // 40 byte (4 * 10 byte)
};
```

size chunk struct dengan pendekatan json: 304 byte (no spaces)
1 json data struct = 25 byte (data paling akhir)
9 json data struct = 26 byte * 9 (ada koma)
total 10 data struct = 259 byte (26 byte * 9 data + 25 byte)

size chunk struct dengan pendekatan json: 228 byte (no spaces & shorten key)
1 json data struct = 19 byte (data paling akhir)
9 json data struct = 20 byte * 9 (ada koma)
total 10 data struct = 199 byte (20 byte * 9 data + 19 byte)

```json
{ // 1 byte
  "timestamp": 1783835980, // 23 byte, 15 byte jika shorten key "t"
  "nodeID": 1, // 11 byte, 6 byte jika shorten key "i"
  "data": [ // 8 byte, 5 byte jika shorten key "d"
    { // 1 byte
      "humi": 3083, // 12 byte, 9 byte jika shorten key "h"
      "temp": 2450 // 11 byte, 8 byte jika shorten key "t"
    }, // 2 byte
    ...
    { // 1 byte
      "humi": 3083, // 12 byte, 9 byte jika shorten key "h"
      "temp": 2450 // 11 byte, 8 byte jika shorten key "t"
    } // 1 byte
  ] // 1 byte
} // 1 byte
```

perbedaan:
- biner: 48 byte
- json (no spaces): 304 byte
- json (no spaces & shorten key): 228 byte

pada level chunk, perbedaan terlihat sangat jelas keunggulan dari pendekatan biner, karena sekalipun pendekatan json di optimasi dengan menghilangkan spasi dan membuat key/kunci nya lebih pendek, perbedaan angka nya cukup jauh yakni 180 byte.

perhitungan ini merupakan langkah awal untuk menghitung sifat compounding dikarenakan sensor akan selalu menyala seharian penuh. sensor akan memproduksi 1 chunk pada 20 detik sekali, pada 1 hari penuh sensor dapat memproduksi sekitar  4320 chunk (didapat dari 86400/20).

jadi, jika kita hitung pada kedua pendekatan diatas, maka didapatkan data berikut:
- biner: 4320 chunk * 48 byte = 207.360 byte
- json (no spaces): 4320 chunk * 304 byte = 1.313.280 byte
- json (no spaces & shorten key): 4320 chunk * 228 byte = 984.960 byte


Analisis Pengaruh Ukuran Payload (X) terhadap Sistem (Y)

Dalam infrastruktur pengiriman data berkelanjutan dari titik node menuju server via protokol MQTT, ukuran payload (X) berdampak langsung dan tidak proporsional pada performa arsitektur (Y):

- Konsumsi Daya (Time-on-Air): Operasi transmisi WiFi pada perangkat edge adalah komponen yang menguras baterai paling agresif. Mentransmisikan paket jaringan berisi 48 byte memerlukan waktu aktif radio (Tx state) yang lebih rendah dibandingkan mengirim 304 byte. Durasi transmisi yang lebih pendek mempercepat perangkat kembali ke kondisi siaga.

- Efisiensi Bandwidth: Protokol MQTT memiliki beban struktur bawaan berupa fixed header dan variable header berisi panjang topik. Memanfaatkan struct biner akan menghasilkan persentase goodput (data aplikatif murni) yang optimal terhadap total besaran lalu lintas jaringan TCP/IP.Stabilitas Memori (Heap): Mengonstruksi kumpulan karakter JSON secara terprogram pada bahasa C/C++ mengharuskan operasi konversi angka menjadi string ASCII yang berulang setiap 20 detik. Praktik manipulasi string ini umumnya memakan alokasi memori dinamis secara masif. Pada periode operasional 24 jam nonstop penuh, proses ini mengundang risiko heap fragmentation hingga kegagalan fungsi (crash).  


Analisis Efek Parsing (Serializzation & Deserialization)

- Sisi Mikrokontroller (Serialization): Dengan menggunakan pendekatan biner, perangkat cukup menggunakan data primitif yaitu chunk dan langsung mengirimkannya menggunakan MQTT. Sedangkan jika menggunakan json, perangkat harus membuat suatu objek baru dengan iterasi objek chunk kembali, yang dimana ini memengaruhi overhead diawal (serialization).

- Sisi Server (Deserialization): Modul encoding/json internal Go memiliki kelemahan struktural karena sangat bergantung pada modul reflection yang membebani kinerja dan memeriksa tipe data secara runtime. Deserialisasi JSON harian akan memproduksi artefak memori secara intensif dan memicu Garbage Collector (GC) bekerja secara konstan. Dengan pendekatan biner, encoding/binary memetakan rantai byte mentah langsung menimpa struct memori Golang tanpa mekanisme penyalinan per karakter.