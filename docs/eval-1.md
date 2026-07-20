# Laporan Evaluasi dan Migrasi Arsitektur IoT Backend

**Tanggal:** 15 Juli 2026
**Pemilik Proyek:** Muhammad Fauzan Faturosi
**Fokus Evaluasi:** Optimalisasi *Ingestion* Data Telemetri & Migrasi ke Ekosistem AWS Serverless

---

## 1. Ringkasan Eksekutif

Laporan ini mengevaluasi arsitektur *backend* pencatatan data lingkungan IoT saat ini yang dibangun menggunakan Golang, Redis FIFO, dan MySQL. Secara tingkat aplikasi (*application layer*), sistem ini memiliki optimasi memori yang sangat tinggi (`0 allocs/op`). Namun, secara infrastruktur dan sistem terdistribusi, arsitektur ini memiliki beberapa leher botol (*bottleneck*) struktural yang membatasi skalabilitas. Laporan ini merekomendasikan transisi menuju arsitektur *cloud-native* terkelola menggunakan AWS IoT Core dan Amazon Timestream untuk memangkas *overhead* operasional (DevOps) dan meningkatkan reliabilitas *ingestion* data.

---

## 2. Analisis Kritis Arsitektur Saat Ini (Legacy System)

Berdasarkan tinjauan *First Principles*, arsitektur saat ini memiliki tiga kerentanan fundamental:

* **Penguncian Pesimis (Pessimistic Locking) MySQL:** Mekanisme penetapan *Node ID* yang bergantung pada `FOR UPDATE` di MySQL saat registrasi perangkat merupakan *anti-pattern*. Pada skenario *reconnect* massal (misal: pasca-pemadaman listrik), hal ini akan memicu badai koneksi, menghabiskan *connection pool*, dan berpotensi membuat *database* mati.
* **Redundansi Middleware (Redis sebagai Buffer):** Penggunaan Redis List FIFO murni diimplementasikan untuk melindungi batas penulisan (*write throughput*) MySQL. Hal ini memindahkan risiko dari *bottleneck* I/O menjadi risiko *Out of Memory* (OOM) pada Redis jika *worker* gagal berfungsi. Jika sistem hulu memproduksi data berskala masif, sistem penyimpanan hilir harus memiliki kapasitas penyerapan yang setara tanpa perlu "plester" berupa *buffer* di memori.
* **Integritas Timestamp:** Menimpa *timestamp* di sisi server (`binary.LittleEndian.PutUint32`) mendistorsi waktu aktual dari *edge device*, membuat data analitik rentan terhadap latensi jaringan.

---

## 3. Target Arsitektur AWS (Proposed Architecture)

Migrasi ke ekosistem AWS memungkinkan transisi dari pengelolaan *server* mandiri ke model yang terkelola penuh (*fully managed*), dengan fokus pada dua komponen utama:

1. **AWS IoT Core:** Menggantikan *broker* MQTT internal. IoT Core menangani miliaran koneksi, otentikasi berbasis sertifikat X.509 (mTLS), dan merutekan pesan tanpa perlu mengelola mesin virtual (EC2).
2. **Amazon Timestream:** Menggantikan kombinasi Redis dan MySQL. Timestream adalah *database time-series serverless* yang dirancang secara *native* untuk menelan triliunan *event* per hari. Timestream menghapus kebutuhan akan `IngestionWorker` kustom dan *buffer* antrean.

---

## 4. Opsi Jalur Migrasi

Mengingat mikrokontroler saat ini mengirimkan *payload* biner 48-byte yang sangat teroptimasi, terdapat dua opsi integrasi menuju AWS Timestream:

| Kriteria | Opsi A: 100% Serverless (Ubah ke JSON) | Opsi B: Hybrid Serverless (Pertahankan Biner) |
| --- | --- | --- |
| **Aliran Data** | Edge (JSON) -> IoT Core -> Timestream | Edge (Biner) -> IoT Core -> Lambda (Go) -> Timestream |
| **Penggunaan Resource Edge** | *Overhead bandwidth* & baterai meningkat. | Sangat hemat (*bandwidth* minimal, 48-byte konstan). |
| **Kompleksitas Infrastruktur** | Sangat rendah. Menggunakan AWS IoT Rules secara langsung. | Menengah. Membutuhkan AWS Lambda sebagai *parser* penengah. |
| **Pemanfaatan Kode Lama** | Seluruh kode Go *backend* saat ini dihapus. | Logika *decoding* biner Go lama di- *porting* ke AWS Lambda. |
| **Skalabilitas Ingestion** | Otomatis (*Native Integration*). | Skalabilitas Lambda bergantung pada batas konkurensi regional. |

---

## 5. Kesimpulan & Rekomendasi Tindakan

Mempertahankan EC2, Redis, dan MySQL untuk sistem yang murni bersifat *data ingestion time-series* tidak sejalan dengan *best practices* pengelolaan siklus hidup sistem (*system lifecycle*) modern, karena beban pemeliharaan infrastruktur jauh lebih besar daripada nilai bisnis yang dihasilkan.

**Langkah Selanjutnya:**

1. **Fase Prototipe (Minggu 1):** Buat purwarupa menggunakan **Opsi A** pada 1-2 perangkat keras percobaan untuk memvalidasi performa baca/tulis Amazon Timestream dan memahami skema penyimpanannya (*magnetic vs. memory store*).
2. **Evaluasi Bandwidth (Minggu 2):** Ukur dampak perubahan ke *payload* JSON terhadap konsumsi daya mikrokontroler. Jika dampaknya di luar batas toleransi perangkat keras keras, terapkan **Opsi B** dengan membungkus ulang logika *parsing* Go (`0 allocs/op`) ke dalam AWS Lambda.
3. **Depresiasi Sistem Lama:** Matikan klaster *IngestionWorker* (Go), Redis, dan turunkan kapasitas MySQL lama setelah data historis bermigrasi.