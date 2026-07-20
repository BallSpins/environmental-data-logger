package generator

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/ballspins/environmental-data-logger/backend/internal/model"
)

// GenerateRandomChunk membuat payload ChunkPayload biner acak (48 byte)
func GenerateRandomChunk(nodeID uint8) model.ChunkPayload {
	var chunk model.ChunkPayload
	chunk.Timestamp = uint32(time.Now().Unix())
	chunk.NodeID = nodeID

	for i := 0; i < 10; i++ {
		// Humi antara 30.00% s.d 90.00% (di-scale ke 3000 - 9000)
		chunk.Data[i].Humi = int16(3000 + rand.Intn(6000))
		// Temp antara 15.00C s.d 40.00C (di-scale ke 1500 - 4000)
		chunk.Data[i].Temp = int16(1500 + rand.Intn(2500))
	}
	return chunk
}

// ChunkToBinary mengonversi ChunkPayload ke biner 48-byte
func ChunkToBinary(chunk model.ChunkPayload, dest []byte) {
	binary.LittleEndian.PutUint32(dest[0:4], chunk.Timestamp)
	dest[4] = chunk.NodeID
	dest[5] = chunk.Padding[0]
	dest[6] = chunk.Padding[1]
	dest[7] = chunk.Padding[2]

	for i := 0; i < 10; i++ {
		offset := 8 + i*4
		binary.LittleEndian.PutUint16(dest[offset:offset+2], uint16(chunk.Data[i].Humi))
		binary.LittleEndian.PutUint16(dest[offset+2:offset+4], uint16(chunk.Data[i].Temp))
	}
}

// JSONPayload merepresentasikan data teragregasi yang setara dengan biner 48-byte
type JSONPayload struct {
	Timestamp uint32          `json:"timestamp"`
	NodeID    uint8           `json:"node_id"`
	Data      []JSONSensorLog `json:"data"`
}

type JSONSensorLog struct {
	Humi float32 `json:"humi"`
	Temp float32 `json:"temp"`
}

// ChunkToJSON mengonversi ChunkPayload ke JSON bytes
func ChunkToJSON(chunk model.ChunkPayload) ([]byte, error) {
	jp := JSONPayload{
		Timestamp: chunk.Timestamp,
		NodeID:    chunk.NodeID,
		Data:      make([]JSONSensorLog, 10),
	}
	for i := 0; i < 10; i++ {
		jp.Data[i] = JSONSensorLog{
			Humi: float32(chunk.Data[i].Humi) / 100.0,
			Temp: float32(chunk.Data[i].Temp) / 100.0,
		}
	}
	return json.Marshal(jp)
}

// CompressGzip mengompres data bytes menggunakan gzip untuk perbandingan kompresi
func CompressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(data)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
