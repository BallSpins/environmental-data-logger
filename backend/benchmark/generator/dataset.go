package generator

import (
	"time"

	"github.com/ballspins/environmental-data-logger/backend/internal/model"
)

// GenerateDeterministicDataset menghasilkan dataset SensorDataRow secara deterministik untuk benchmark database
func GenerateDeterministicDataset(numRows int, nodeID uint8) []model.SensorDataRow {
	dataset := make([]model.SensorDataRow, numRows)
	baseTime := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)

	for i := 0; i < numRows; i++ {
		// Buat temperatur & kelembapan yang bergradasi secara deterministik berdasarkan indeks
		tempVal := 20.0 + float32(i%1500)/100.0
		humiVal := 50.0 + float32(i%3000)/100.0

		// Mundur per baris 2 detik secara teratur
		exactTime := baseTime.Add(time.Duration(-i*2) * time.Second)

		dataset[i] = model.SensorDataRow{
			Timestamp: exactTime,
			NodeID:    nodeID,
			Temp:      tempVal,
			Humi:      humiVal,
		}
	}
	return dataset
}
