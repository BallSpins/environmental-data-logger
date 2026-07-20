#pragma once

/// Peripheral config
#define DHT_PIN 32
#define CS_PIN 27
#define SPI_MOSI 26
#define SPI_MISO 25
#define SPI_SCK  33

static const char* wifi_ssid = "Wokwi-GUEST";
static const char* wifi_pass = "";
static const char* mqtt_server = "d21522690eb14434b34e54db4c2890dd.s1.eu.hivemq.cloud"; // Sesuaikan dengan broker Anda
static const char* mqtt_user = "sensor-1"; // Ganti dengan user yang Anda buat di dasbor HiveMQ
static const char* mqtt_pass = "sensor-1";
static const char* registerTopic = "devices/register";

const int   mqtt_port = 8883; // Gunakan port TLS sesuai dasbor