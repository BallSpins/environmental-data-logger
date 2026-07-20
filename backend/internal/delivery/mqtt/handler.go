package mqtt

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ballspins/environmental-data-logger/backend/internal/repository"
	"github.com/ballspins/environmental-data-logger/backend/internal/service"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTHandler struct {
	deviceRepo *repository.DeviceRepository
	cacheRepo  *repository.CacheRepository
	parser     *service.ParserService
}

func NewMQTTHandler(
	dr *repository.DeviceRepository,
	cr *repository.CacheRepository,
	ps *service.ParserService,
) *MQTTHandler {
	return &MQTTHandler{
		deviceRepo: dr,
		cacheRepo:  cr,
		parser:     ps,
	}
}

// HandleRegister memproses pendaftaran MAC Address menggunakan ACID Transaction
func (h *MQTTHandler) HandleRegister(client mqtt.Client, msg mqtt.Message) {
	mac := string(msg.Payload())
	log.Printf("[MQTT Rx] Menerima request registrasi untuk MAC: %s\n", mac)

	// Panggil repository transaksional (ACID) untuk mendapatkan/membuat ID
	nodeID, err := h.deviceRepo.RegisterOrGetID(mac)
	if err != nil {
		log.Printf("[ERROR] Gagal registrasi device %s: %v\n", mac, err)
		return
	}

	// Kirim balik ID ke topik assignment khusus perangkat tersebut
	responseTopic := fmt.Sprintf("devices/%s/assignment", mac)
	payloadStr := strconv.Itoa(int(nodeID))

	token := client.Publish(responseTopic, 1, false, payloadStr)
	token.Wait()

	log.Printf("[MQTT Tx] Berhasil mengalokasikan Node ID %d ke topik %s\n", nodeID, responseTopic)
}

// HandleSensorData memproses payload biner 48-byte secara optimal dan melemparkannya ke Redis Buffer tanpa alokasi heap
func (h *MQTTHandler) HandleSensorData(client mqtt.Client, msg mqtt.Message) {
	payload := msg.Payload()

	// 1. Validasi integritas ukuran paket biner (Wajib 48 byte)
	if len(payload) != 48 {
		log.Printf("[ERROR] Payload biner invalid size dari topik %s: expected 48 bytes, got %d\n", msg.Topic(), len(payload))
		return
	}

	// 2. Injeksi Unix timestamp detik saat ini ke dalam 4 byte pertama payload secara Little Endian tanpa alokasi heap
	binary.LittleEndian.PutUint32(payload[0:4], uint32(time.Now().Unix()))

	// 3. Lempar raw payload biner ke Redis List Buffer secara langsung (BASE Layer)
	ctx := context.Background()
	err := h.cacheRepo.BufferSensorData(ctx, payload)
	if err != nil {
		log.Printf("[ERROR] Gagal menyimpan data biner ke Redis Buffer: %v\n", err)
		return
	}

	// Ambil NodeID dari byte ke-4 untuk logging tanpa alokasi
	nodeID := payload[4]
	log.Printf("[Ingestion] Berhasil mem-buffer data biner dari Node ID %d ke Redis.\n", nodeID)
}
