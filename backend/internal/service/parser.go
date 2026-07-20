package service

import (
	"encoding/binary"
	"fmt"

	"github.com/ballspins/environmental-data-logger/backend/internal/model"
)

type ParserService struct{}

func NewParserService() *ParserService {
	return &ParserService{}
}

// ParseBinary membedah raw bytes dari MQTT langsung ke dest pointer dengan performa O(1) dan 0 alokasi heap
func (s *ParserService) ParseBinary(payload []byte, dest *model.ChunkPayload) error {
	// Validasi integritas ukuran paket (Wajib 48 byte)
	if len(payload) != 48 {
		return fmt.Errorf("invalid payload size: expected 48 bytes, got %d", len(payload))
	}

	// Membaca biner mentah secara manual ke struct tujuan (Little Endian) - 100% aman, cepat, dan 0 alokasi
	dest.Timestamp = binary.LittleEndian.Uint32(payload[0:4])
	dest.NodeID = payload[4]
	dest.Padding[0] = payload[5]
	dest.Padding[1] = payload[6]
	dest.Padding[2] = payload[7]

	for i := 0; i < 10; i++ {
		offset := 8 + i*4
		dest.Data[i].Humi = int16(binary.LittleEndian.Uint16(payload[offset : offset+2]))
		dest.Data[i].Temp = int16(binary.LittleEndian.Uint16(payload[offset+2 : offset+4]))
	}

	return nil
}
