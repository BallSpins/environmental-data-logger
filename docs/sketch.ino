#include <DHT.h>
#include <SD.h>
#include <SPI.h>
#include <WiFi.h>
#include <Wire.h>
#include <RTClib.h>
#include <WiFiClientSecure.h>
#include <PubSubClient.h>
#include <Preferences.h>

/// Peripheral config
#define DHT_PIN 32
#define CS_PIN 27
#define SPI_MOSI 26
#define SPI_MISO 25
#define SPI_SCK 33

// Wi-Fi & MQTT config
const char *ssid = "Wokwi-GUEST";
const char *password = "";
const char *mqtt_server = "d21522690eb14434b34e54db4c2890dd.s1.eu.hivemq.cloud"; // Sesuaikan dengan broker Anda
const int mqtt_port = 8883;                                                      // Gunakan port TLS sesuai dasbor
const char *mqtt_user = "sensor-1";                                              // Ganti dengan user yang Anda buat di dasbor HiveMQ
const char *mqtt_pass = "sensor-1";
const char *mqtt_topic = "esp32/sensor/chunk";

RTC_DS1307 rtc;
DHT dht22(DHT_PIN, DHT22);
WiFiClientSecure espClientSecure;
PubSubClient mqttClient(espClientSecure);

// Data struct (4 bytes)
struct Data
{
  int16_t humi;
  int16_t temp;
};

// Chunk struct 42 bytes
struct Chunk
{
  uint32_t timestamp;
  uint8_t nodeID;
  uint8_t padding[3];
  Data data[10]; // 4 * 10 bytes
};

// Handle FreeRTOS Queue
QueueHandle_t dataQueue;
Preferences prefs;

// Task Function Declaration
void TaskSensorCore1(void *pvParameters);
void TaskCommCore0(void *pvParameters);

uint8_t myNodeID = 0; // 0 means didnt registered yet
bool isRegistered = false;
const char *registerTopic = "devices/register";
char assignmentTopic[36];
char dataTopic[36];

void setup()
{
  // put your setup code here, to run once:
  Serial.begin(115200);

  prefs.begin("dev_config", false);

  myNodeID = prefs.getUChar("node_id", 0);

  uint8_t mac[6];
  WiFi.macAddress(mac);

  snprintf(assignmentTopic, sizeof(assignmentTopic), "devices/%02X%02X%02X%02X%02X%02X/assignment",
           mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);

  snprintf(dataTopic, sizeof(dataTopic), "devices/%d/data", myNodeID);

  if (myNodeID == 0)
  {
    Serial.println("[System] Perangkat BELUM teregistrasi. Menunggu ID dari server...");
    isRegistered = false;
  }
  else
  {
    Serial.printf("[System] Perangkat SUDAH teregistrasi. Menggunakan Node ID: %d\n", myNodeID);
    isRegistered = true;
  }

  Wire.begin(14, 13);
  if (!rtc.begin())
  {
    Serial.println("[System] Modul RTC tidak terdeteksi!");
    while (1)
      ;
  }

  // if (rtc.lostPower()) {
  //   Serial.println("[System] RTC kehilangan daya, mengatur waktu kompilasi...");
  //   // Ini hanya untuk darurat jika baterai kancing habis
  //   rtc.adjust(DateTime(F(__DATE__), F(__TIME__)));
  // }

  dht22.begin();

  // Init SPI & SD Card
  SPI.begin(SPI_SCK, SPI_MISO, SPI_MOSI, CS_PIN);
  if (!SD.begin(CS_PIN, SPI))
  {
    Serial.println("Card initialization failed!");
    while (true)
      ;
  }
  Serial.println("SD Card Berhasil diinisialisasi.");

  // Init FreeRTOS Queue for 5 chunks object (5 * 40 bytes)
  dataQueue = xQueueCreate(5, sizeof(Chunk));
  if (dataQueue == NULL)
  {
    Serial.println("Gagal membuat Queue!");
    while (true)
      ;
  }

  mqttClient.setCallback(mqttCallback);

  // Spawning Task to Core 1 (Sensor)
  xTaskCreatePinnedToCore(
      TaskSensorCore1, // Task function
      "TaskSensor",    // Task name
      4000,            // Stack size
      NULL,            // Parameter
      2,               // Priority lebih tinggi agar sampling tepat waktu
      NULL,            // Handle
      1                // Runs on Core 1
  );

  // Spawning Task to Core 0 (Wi-Fi, MQTT, & SD Card Storage)
  xTaskCreatePinnedToCore(
      TaskCommCore0, // Task function
      "TaskComm",    // Task name
      8000,          // Stack size (lebih besar karena handle Wi-Fi/IP stack)
      NULL,          // Parameter
      1,             // Priority
      NULL,          // Handle
      0              // Runs on Core 0
  );
}

void loop()
{
  // put your main code here, to run repeatedly:
  vTaskDelete(NULL);
}

// ==========================================
// CORE 1: TASK PENGAMBILAN DATA SENSOR
// ==========================================
void TaskSensorCore1(void *pvParameters)
{
  Chunk localChunk;
  int index = 0;

  for (;;)
  {
    vTaskDelay(2000 / portTICK_PERIOD_MS); // Sampling setiap 2 detik

    float humi = dht22.readHumidity();
    float tempC = dht22.readTemperature();

    if (isnan(humi) || isnan(tempC))
    {
      Serial.println("[Core 1] Gagal membaca sensor DHT22!");
      continue;
    }

    // Kompresi data
    localChunk.data[index].humi = (int16_t)(humi * 100.0f);
    localChunk.data[index].temp = (int16_t)(tempC * 100.0f);

    Serial.printf("[Core 1] Slot %d terisi. Temp: %.2f\n", index, tempC);
    index++;

    if (index >= 10)
    {
      Serial.println("[Core 1] Chunk penuh! Mengirim ke Queue...");
      DateTime now = rtc.now();
      localChunk.timestamp = now.unixtime();

      // Kirim salinan mentah localChunk ke dalam Queue.
      // Jika queue penuh, tunggu maksimal 100ms sebelum drop.
      if (xQueueSend(dataQueue, &localChunk, 100 / portTICK_PERIOD_MS) != pdPASS)
      {
        Serial.println("[Core 1] Queue penuh! Data terpaksa di-drop.");
      }

      index = 0; // Reset index untuk siklus berikutnya
    }
  }
}

// ==========================================
// CORE 0: TASK WI-FI, MQTT, & STORAGE (FIFO)
// ==========================================
void TaskCommCore0(void *pvParameters)
{
  Chunk rxChunk;

  // Inisialisasi Koneksi awal
  WiFi.begin(ssid, password);
  espClientSecure.setInsecure();
  mqttClient.setServer(mqtt_server, mqtt_port);

  for (;;)
  {
    // 1. Jaga konektivitas Wi-Fi & MQTT secara non-blocking
    if (WiFi.status() != WL_CONNECTED)
    {
      vTaskDelay(500 / portTICK_PERIOD_MS);
      continue;
    }
    if (!mqttClient.connected())
    {
      Serial.println("[Core 0] Membuka jabat tangan TLS ke HiveMQ Cloud...");

      uint8_t mac[6];
      WiFi.macAddress(mac);
      char clientID[30];
      snprintf(clientID, sizeof(clientID), "ESP32_%02X%02X%02X%02X%02X%02X",
               mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);

      if (mqttClient.connect(clientID, mqtt_user, mqtt_pass))
      {
        Serial.println("[Core 0] Koneksi aman MQTT Berhasil Terhubung!");

        if (!isRegistered)
        {
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
      }
      else
      {
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

    if (isRegistered)
    {
      if (mqttClient.connected() && SD.exists("/datalog.bin"))
      {
        processSDCardQueue();
      }

      if (xQueueReceive(dataQueue, &rxChunk, 10 / portTICK_PERIOD_MS) == pdPASS)
      {

        // Wajib set identitas biner sebelum masuk antrean pengiriman / SD Card
        rxChunk.nodeID = myNodeID;

        bool sdCardHasData = SD.exists("/datalog.bin");
        bool mqttReady = mqttClient.connected();

        // LOGIKA FIFO SINGLE PIPE (Tetap Sama)
        if (sdCardHasData)
        {
          writeChunkToSD(rxChunk);
          if (mqttReady)
            processSDCardQueue();
        }
        else
        {
          if (mqttReady && sendChunkToMQTT(rxChunk))
          {
            Serial.println("[Core 0] Kirim langsung lewat MQTT Berhasil.");
          }
          else
          {
            writeChunkToSD(rxChunk);
          }
        }
      }
    }
  }
}

// ==========================================
// FUNGSI-FUNGSI HELPER I/O
// ==========================================

bool sendChunkToMQTT(Chunk &chunk)
{
  // 1. Ambil pointer biner dari struct Chunk (40 bytes)
  uint8_t *binaryPayload = (uint8_t *)&chunk;
  size_t payloadLength = sizeof(Chunk); // Pasti 40 bytes

  snprintf(dataTopic, sizeof(dataTopic), "devices/%d/data", chunk.nodeID);

  // 2. Kirim biner mentah langsung ke broker MQTT
  // Parameter: (topic, payload_pointer, length)
  return mqttClient.publish(dataTopic, binaryPayload, payloadLength);
}

void writeChunkToSD(Chunk &chunk)
{
  File dFile = SD.open("/datalog.bin", FILE_APPEND);
  if (dFile)
  {
    dFile.write((uint8_t *)&chunk, sizeof(Chunk));
    dFile.close();
    Serial.println("[Storage] Append chunk ke SD Card sukses.");
  }
  else
  {
    Serial.println("[Storage] Gagal membuka file untuk menyimpan.");
  }
}

void processSDCardQueue()
{
  File dataFile = SD.open("/datalog.bin", FILE_READ);
  if (!dataFile)
    return;

  Serial.println("[Storage] Menguras data antrean dari SD Card...");

  Chunk tempChunk;
  bool allSentSuccessfully = true;

  // Baca satu per satu chunk dari yang paling lama (FIFO)
  while (dataFile.read((uint8_t *)&tempChunk, sizeof(Chunk)) == sizeof(Chunk))
  {
    if (sendChunkToMQTT(tempChunk))
    {
      Serial.println("[Storage] Satu data antrean sukses dikirim ke MQTT.");
    }
    else
    {
      Serial.println("[Storage] Pengiriman antrean terhenti (MQTT putus lagi).");
      allSentSuccessfully = false;
      break;
    }
  }
  dataFile.close();

  // Jika semua data historis sukses didepak ke MQTT, hapus file untuk reset antrean
  if (allSentSuccessfully)
  {
    SD.remove("/datalog.bin");
    Serial.println("[Storage] Seluruh antrean bersih. File datalog.bin dihapus.");
  }
  else
  {
    // Catatan Optimasi Lanjutan: Jika gagal di tengah jalan, idealnya Anda memotong
    // bagian file yang sudah terkirim. Namun untuk simulasi dasar, file dibiarkan
    // dan akan dicoba ulang pada siklus pengiriman berikutnya.
  }
}

void mqttCallback(char *topic, byte *payload, unsigned int length)
{
  // Pastikan pesan datang dari topik assignment perangkat ini
  Serial.printf("[MQTT DEBUG] Menerima pesan di topik: %s\n", topic);
  if (String(topic) == assignmentTopic && !isRegistered)
  {
    char idStr[length + 1];
    memcpy(idStr, payload, length);
    idStr[length] = '\0';

    int assignedID = atoi(idStr);
    if (assignedID > 0 && assignedID <= 255)
    {
      myNodeID = (uint8_t)assignedID;

      // BAKAR KE FLASH MEMORI (NVS)
      prefs.putUChar("node_id", myNodeID);
      isRegistered = true;

      Serial.printf("[MQTT] BERHASIL! Server memberikan Node ID: %d. Tersimpan permanen.\n", myNodeID);
    }
  }
}
