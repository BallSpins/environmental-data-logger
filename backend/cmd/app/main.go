package main

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/ballspins/environmental-data-logger/backend/internal/config"
	"github.com/ballspins/environmental-data-logger/backend/internal/model"
	"github.com/ballspins/environmental-data-logger/backend/internal/repository"
	"github.com/ballspins/environmental-data-logger/backend/internal/service"
	"github.com/ballspins/environmental-data-logger/backend/internal/worker"
	mqtt "github.com/eclipse/paho.mqtt.golang"

	deliveryMQTT "github.com/ballspins/environmental-data-logger/backend/internal/delivery/mqtt"
)

func main() {
	log.Println("[System] Memulai inisialisasi backend data logger...")

	// 1. Load Konfigurasi dari .env
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[Config] Gagal memuat .env: %v", err)
	}
	log.Println("[Config] File .env berhasil dimuat.")

	db, sqlDB, err := config.InitMysql(cfg)

	if err != nil {
		log.Fatalf("[MySQL] Koneksi gagal dibuat: %v", err)
	}

	err = db.AutoMigrate(
		&model.RegisteredDevice{},
		&model.SensorDataRow{},
	)
	if err != nil {
		log.Fatalf("[MySQL] Gagal melakukan AutoMigrate tabel: %v", err)
	}
	log.Println("[MySQL] AutoMigrate berhasil. Struktur tabel terverifikasi.")

	rdb, err := config.InitRedis(cfg)
	if err != nil {
		log.Fatalf("[Redis] Gagal terhubung ke instansi Redis: %v", err)
	}

	deviceRepo := repository.NewDeviceRepository(db)
	cacheRepo := repository.NewCacheRepository(rdb)
	sensorRepo := repository.NewSensorRepository(sqlDB)
	parserService := service.NewParserService()

	ingestionWorker := worker.NewIngestionWorker(rdb, sensorRepo, parserService)
	ingestionWorker.Start()

	mqttHandler := deliveryMQTT.NewMQTTHandler(deviceRepo, cacheRepo, parserService)

	opts := mqtt.NewClientOptions()
	brokerURI := fmt.Sprintf("tls://%s:%s", cfg.MQTTHost, cfg.MQTTPort)
	opts.AddBroker(brokerURI)

	opts.SetClientID("go_backend_environmental_logger")
	opts.SetCleanSession(true)

	opts.SetUsername(cfg.MQTTUser)
	opts.SetPassword(cfg.MQTTPass)

	opts.SetTLSConfig(&tls.Config{
		MinVersion: tls.VersionTLS12,
	})

	opts.OnConnect = func(client mqtt.Client) {
		log.Println("[MQTT] Berhasil terhubung ke broker.")

		// Subscribe: Registrasi
		client.Subscribe("devices/register", 1, mqttHandler.HandleRegister)
		log.Println("[MQTT] Mendengarkan topik: devices/register")

		// Subscribe: Data Sensor
		client.Subscribe("devices/+/data", 1, mqttHandler.HandleSensorData)
		log.Println("[MQTT] Mendengarkan topik: devices/+/data")
	}

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("[MQTT] Gagal terhubung ke broker: %v", token.Error())
	}

	// Log status siap pakai
	log.Println("[System] Seluruh infrastruktur data (MySQL & Redis) siap digunakan.")

	select {}
}
