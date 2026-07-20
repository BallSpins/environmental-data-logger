#include <Arduino.h>
#include <DHT.h>
#include <SD.h>
#include <SPI.h>
#include <Wire.h>
#include <RTClib.h>
#include <Preferences.h>

#include "Config.h"
#include "Types.h"
#include "Globals.h"
#include "Network.h"
#include "Storage.h"
#include "Filter.h"

QueueHandle_t dataQueue;
Preferences prefs;
uint8_t myNodeID = 0;
bool isRegistered = false;
char assignmentTopic[36];
char dataTopic[36];

RTC_DS1307 rtc;
DHT dht22(DHT_PIN, DHT22);

EMAFilter tempFilter(0.1f);
EMAFilter humiFilter(0.1f);

void TaskSensorCore1(void *pvParameters);
void TaskCommCore0(void *pvParameters);

void setup() {
  Serial.begin(115200);

  prefs.begin("dev_config", false);
  myNodeID = prefs.getUChar("node_id", 0);

  uint8_t mac[6];
  WiFi.macAddress(mac);
  snprintf(assignmentTopic, sizeof(assignmentTopic), "devices/%02X%02X%02X%02X%02X%02X/assignment", 
           mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);
  snprintf(dataTopic, sizeof(dataTopic), "devices/%d/data", myNodeID);

  if (myNodeID == 0) {
    Serial.println("[System] Perangkat BELUM teregistrasi. Menunggu ID dari server...");
    isRegistered = false;
  } else {
    Serial.printf("[System] Perangkat SUDAH teregistrasi. Menggunakan Node ID: %d\n", myNodeID);
    isRegistered = true;
  }

  Wire.begin(14, 13);
  if (!rtc.begin()) {
    Serial.println("[System] Modul RTC tidak terdeteksi!");
    while (1); 
  }

  dht22.begin();

  SPI.begin(SPI_SCK, SPI_MISO, SPI_MOSI, CS_PIN);
  if (!SD.begin(CS_PIN, SPI)) {
    Serial.println("Card initialization failed!");
    while (true);
  }
  Serial.println("SD Card Berhasil diinisialisasi.");

  dataQueue = xQueueCreate(5, sizeof(Chunk));
  if (dataQueue == NULL) {
    Serial.println("Gagal membuat Queue!");
    while (true);
  }

  initNetwork();

  // Spawning Task to Core 1 (Sensor)
  xTaskCreatePinnedToCore(
    TaskSensorCore1,   // Task function
    "TaskSensor",      // Task name
    4000,              // Stack size
    NULL,              // Parameter
    2,                 // Priority lebih tinggi agar sampling tepat waktu
    NULL,              // Handle
    1                  // Runs on Core 1
  );

  // Spawning Task to Core 0 (Wi-Fi, MQTT, & SD Card Storage)
  xTaskCreatePinnedToCore(
    TaskCommCore0,     // Task function
    "TaskComm",        // Task name
    8000,              // Stack size (lebih besar karena handle Wi-Fi/IP stack)
    NULL,              // Parameter
    1,                 // Priority
    NULL,              // Handle
    0                  // Runs on Core 0
  );
}

void loop() {
  vTaskDelete(NULL);
}

// ==========================================
// CORE 1: TASK PENGAMBILAN DATA SENSOR
// ==========================================
void TaskSensorCore1(void *pvParameters) {
  Chunk localChunk;
  int index = 0;

  for (;;) {
    vTaskDelay(2000 / portTICK_PERIOD_MS); // Sampling setiap 2 detik

    float humi  = dht22.readHumidity();
    float tempC = dht22.readTemperature();

    if (isnan(humi) || isnan(tempC)) {
      Serial.println("[Core 1] Gagal membaca sensor DHT22!");
      continue;
    }

    // Filter data
    float filteredHumi = humiFilter(humi);
    float filteredTemp = tempFilter(tempC);

    // Kompresi data
    localChunk.data[index].humi = (int16_t)(filteredHumi * 100.0f);
    localChunk.data[index].temp = (int16_t)(filteredTemp * 100.0f);
    
    Serial.printf("[Core 1] Slot %d terisi. Temp: %.2f\n", index, tempC);
    index++;

    if (index >= 10) {
      Serial.println("[Core 1] Chunk penuh! Mengirim ke Queue...");
      DateTime now = rtc.now();
      localChunk.timestamp = now.unixtime();

      // Kirim salinan mentah localChunk ke dalam Queue. 
      // Jika queue penuh, tunggu maksimal 100ms sebelum drop.
      if (xQueueSend(dataQueue, &localChunk, 100 / portTICK_PERIOD_MS) != pdPASS) {
        Serial.println("[Core 1] Queue penuh! Data terpaksa di-drop.");
      }
      
      index = 0; // Reset index untuk siklus berikutnya
    }
  }
}

// ==========================================
// CORE 0: TASK WI-FI, MQTT, & STORAGE (FIFO)
// ==========================================
void TaskCommCore0(void *pvParameters) {
  Chunk rxChunk;
  
  // Inisialisasi Koneksi awal
  // WiFi.begin(ssid, password);
  // espClientSecure.setInsecure();
  // mqttClient.setServer(mqtt_server, mqtt_port);

  for (;;) {
    // 1. Jaga konektivitas Wi-Fi & MQTT secara non-blocking
    if (WiFi.status() != WL_CONNECTED) {
      vTaskDelay(500 / portTICK_PERIOD_MS);
      continue; 
    }
    if (!mqttClient.connected()) {
      Serial.println("[Core 0] Membuka jabat tangan TLS ke HiveMQ Cloud...");

      uint8_t mac[6];
      WiFi.macAddress(mac);
      char clientID[30];
      snprintf(clientID, sizeof(clientID), "ESP32_%02X%02X%02X%02X%02X%02X", 
               mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);

      if (mqttClient.connect(clientID, mqtt_user, mqtt_pass)) {
        Serial.println("[Core 0] Koneksi aman MQTT Berhasil Terhubung!");

        if (!isRegistered) {
          uint8_t mac[6];
          WiFi.macAddress(mac);
          snprintf(assignmentTopic, sizeof(assignmentTopic), "devices/%02X%02X%02X%02X%02X%02X/assignment", 
                   mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);

          mqttClient.subscribe(assignmentTopic);
          Serial.println("[Core 0] Subscribed ke topik assignment.");

          char macStr[13];
          snprintf(macStr, sizeof(macStr), "%02X%02X%02X%02X%02X%02X", 
                   mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);

          mqttClient.publish(registerTopic, macStr);
          Serial.printf("[Core 0] Meminta ID dari Server untuk MAC: %s\n", macStr);
        }
      } else {
        Serial.printf("[Core 0] Gagal terhubung ke Broker, RC = %d\n", mqttClient.state());
        
        // KRUSIAL: Jika gagal, wajib beri vTaskDelay di sini agar Core 0 
        // tidak langsung menghantam loop berikutnya secara brutal.
        vTaskDelay(5000 / portTICK_PERIOD_MS); 
        continue; // Kembali ke awal loop setelah 5 detik
      }
    
    }
    mqttClient.loop();
    vTaskDelay(10 / portTICK_PERIOD_MS);

    // 2. Ambil data dari Queue jika dikirim oleh Core 1 (Membendung data)
    // portMAX_DELAY membuat task ini tertidur (suspending) jika queue kosong,
    // sehingga tidak memakan siklus CPU Core 0 secara sia-sia.

    if (isRegistered) {
      if (mqttClient.connected() && SD.exists("/datalog.bin")) {
        processSDCardQueue();
      }
      
      if (xQueueReceive(dataQueue, &rxChunk, 10 / portTICK_PERIOD_MS) == pdPASS) {
        
        // Wajib set identitas biner sebelum masuk antrean pengiriman / SD Card
        rxChunk.nodeID = myNodeID;
        
        bool sdCardHasData = SD.exists("/datalog.bin");
        bool mqttReady = mqttClient.connected();

        // LOGIKA FIFO SINGLE PIPE (Tetap Sama)
        if (sdCardHasData) {
          writeChunkToSD(rxChunk);
          if (mqttReady) processSDCardQueue();
        } else {
          if (mqttReady && sendChunkToMQTT(rxChunk)) {
            Serial.println("[Core 0] Kirim langsung lewat MQTT Berhasil.");
          } else {
            writeChunkToSD(rxChunk);
          }
        }
      }
    }
  }
}