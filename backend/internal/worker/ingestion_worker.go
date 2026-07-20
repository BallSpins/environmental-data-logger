package worker

import (
	"context"
	"log"
	"time"
	"unsafe"

	"github.com/ballspins/environmental-data-logger/backend/internal/model"
	"github.com/ballspins/environmental-data-logger/backend/internal/repository"
	"github.com/ballspins/environmental-data-logger/backend/internal/service"
	"github.com/redis/go-redis/v9"
)

type IngestionWorker struct {
	rdb        *redis.Client
	sensorRepo *repository.SensorRepository
	parser     *service.ParserService
	dbRowsBuf  []model.SensorDataRow
	resultsBuf []string
}

func NewIngestionWorker(
	rdb *redis.Client,
	sensorRepo *repository.SensorRepository,
	parser *service.ParserService,
) *IngestionWorker {
	return &IngestionWorker{
		rdb:        rdb,
		sensorRepo: sensorRepo,
		parser:     parser,
		dbRowsBuf:  make([]model.SensorDataRow, 0, 1000), // Pre-alloc untuk menghindari GC pressure selama pemrosesan batch
		resultsBuf: make([]string, 0, 100),               // Pre-alloc untuk fallback loop RPOP jika diperlukan
	}
}

func (w *IngestionWorker) Start() {
	go func() {
		for {
			w.drainRedisToDB()
			time.Sleep(5 * time.Second) // Jeda antar batch untuk mencegah CPU throttling
		}
	}()
}

func (w *IngestionWorker) drainRedisToDB() {
	ctx := context.Background()
	queueKey := "queue:sensor_ingestion"
	batchSize := 100

	// 1. Ambil batch dari Redis menggunakan RPopCount (Sangat cepat & 1 roundtrip network)
	results, err := w.rdb.RPopCount(ctx, queueKey, batchSize).Result()

	if err != nil && err != redis.Nil {
		// Fallback ke loop RPop individu jika versi Redis lama tidak mendukung RPopCount
		results = w.resultsBuf[:0]
		for i := 0; i < batchSize; i++ {
			res, err := w.rdb.RPop(ctx, queueKey).Result()
			if err == redis.Nil {
				break
			} else if err != nil {
				log.Printf("[Worker] Gagal RPOP dari Redis: %v\n", err)
				break
			}
			results = append(results, res)
		}
		w.resultsBuf = results // Simpan kapasitas slice hasil buffering
	}

	if len(results) == 0 {
		return // Antrean kosong
	}

	// Reset slice buffer DB rows ke panjang 0, mempertahankan kapasitas yang telah dialokasikan (0 allocations)
	w.dbRowsBuf = w.dbRowsBuf[:0]

	for _, rawStr := range results {
		// Payload biner dari device WAJIB 48 byte
		if len(rawStr) != 48 {
			log.Printf("[Worker] Payload invalid length di Redis (%d bytes), data di-drop.\n", len(rawStr))
			continue
		}

		// Konversi string ke []byte secara in-place dengan 0 alokasi heap
		payloadBytes := unsafe.Slice(unsafe.StringData(rawStr), len(rawStr))

		// Dekode binary data langsung ke struct stack-allocated
		var chunk model.ChunkPayload
		if err := w.parser.ParseBinary(payloadBytes, &chunk); err != nil {
			log.Printf("[Worker] Gagal parsing payload biner: %v\n", err)
			continue
		}

		baseTime := time.Unix(int64(chunk.Timestamp), 0)

		// Flattening: Memecah array [10]Data dari ESP32 menjadi baris-baris individual untuk DB
		for i := 0; i < 10; i++ {
			dp := chunk.Data[i]
			// Mundur 2 detik untuk setiap indeks array agar data time-series berurutan presisi
			exactTime := baseTime.Add(time.Duration(-(9-i)*2) * time.Second)

			w.dbRowsBuf = append(w.dbRowsBuf, model.SensorDataRow{
				Timestamp: exactTime,
				NodeID:    chunk.NodeID,
				Temp:      float32(dp.Temp) / 100.0, // Konversi balik dari int16 ke float desimal asli
				Humi:      float32(dp.Humi) / 100.0,
			})
		}
	}

	// 2. Tulis batch data sensor ke database menggunakan raw batch-insert zero-allocation
	err = w.sensorRepo.BatchInsert(ctx, w.dbRowsBuf)
	if err != nil {
		// Mekanisme Dead Letter Queue: Jika DB mati, kembalikan data ke Redis agar tidak hilang
		log.Printf("[Worker] Gagal Batch Insert ke DB: %v. Mengembalikan data ke antrean.\n", err)
		for _, rawStr := range results {
			w.rdb.RPush(ctx, queueKey, rawStr)
		}
		return
	}

	log.Printf("[Worker] Sukses memindahkan %d baris data sensor (dari %d chunk biner) ke DB Persisten.\n", len(w.dbRowsBuf), len(results))
}
