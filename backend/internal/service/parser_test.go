package service

import (
	"testing"

	"github.com/ballspins/environmental-data-logger/backend/internal/model"
)

func TestParseBinary(t *testing.T) {
	parser := NewParserService()

	// Simulasi payload biner 48 byte seson data
	// Byte 0-3: Timestamp = 1700000000 (0x65561B00)
	// Byte 4: Node ID = 45
	// Byte 5-7: Reserved Padding = [0, 0, 0]
	// Byte 8-9: Humi ke-1 = 6000 (Artinya 60.00%) -> Dalam biner Little Endian: 0x70, 0x17
	// Byte 10-11: Temp ke-1 = 2550 (Artinya 25.50°) -> Dalam biner Little Endian: 0xF6, 0x09
	// Sisa byte diisi 0 untuk menyederhanakan test slot 2-10
	mockPayload := make([]byte, 48)

	mockPayload[0] = 0x00
	mockPayload[1] = 0x1B
	mockPayload[2] = 0x56
	mockPayload[3] = 0x65

	mockPayload[4] = 45 // Node ID

	// Slot 0 Humi (6000 = 0x1770)
	mockPayload[8] = 0x70
	mockPayload[9] = 0x17
	// Slot 0 Temp (2550 = 0x09F6)
	mockPayload[10] = 0xF6
	mockPayload[11] = 0x09

	var chunk model.ChunkPayload
	err := parser.ParseBinary(mockPayload, &chunk)
	if err != nil {
		t.Fatalf("Gagal melakukan parsing: %v", err)
	}

	// Validasi Node ID & Timestamp
	if chunk.NodeID != 45 {
		t.Errorf("Ekspektasi NodeID 45, didapat %d", chunk.NodeID)
	}
	if chunk.Timestamp != 1700141824 {
		t.Errorf("Ekspektasi Timestamp 1700141824, didapat %d", chunk.Timestamp)
	}

	// Validasi Nilai Sensor Pertama setelah dikonversi kembali ke desimal
	humiDesimal := float64(chunk.Data[0].Humi) / 100.0
	tempDesimal := float64(chunk.Data[0].Temp) / 100.0

	if humiDesimal != 60.00 {
		t.Errorf("Ekspektasi Kelembapan 60.00, didapat %.2f", humiDesimal)
	}
	if tempDesimal != 25.50 {
		t.Errorf("Ekspektasi Suhu 25.50, didapat %.2f", tempDesimal)
	}
}

func BenchmarkParseBinary(b *testing.B) {
	parser := NewParserService()
	mockPayload := make([]byte, 48)
	mockPayload[4] = 45
	mockPayload[8] = 0x70
	mockPayload[9] = 0x17
	mockPayload[10] = 0xF6
	mockPayload[11] = 0x09

	var chunk model.ChunkPayload

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = parser.ParseBinary(mockPayload, &chunk)
	}
}
