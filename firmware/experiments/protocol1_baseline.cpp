#include <Arduino.h>
#include <Wire.h>
#include <WiFi.h>
#include <Adafruit_AHTX0.h>

#include <Filter.h>

Adafruit_AHTX0 aht;

EMAFilter tempFilter(0.1494f);
EMAFilter humiFilter(0.226f);

// Deklarasi Task
void TaskAcquisitionCore0(void *pvParameters);

void setup() {
  Serial.begin(115200);

  // Matikan semua fungsi radio untuk mencegah Thermal Interference
  WiFi.mode(WIFI_OFF);
  btStop(); 

  // Inisialisasi Sensor AHT10
  if (!aht.begin()) {
    Serial.println("Gagal menemukan AHT10. Periksa kabel!");
    while (1) delay(10);
  }

  // Header CSV
  Serial.println("timestamp_ms,temperature_c,humidity_rh,ema_temperature_c,ema_humidity_rh");

  // Spawning Task murni di Core 0 untuk Deterministic Sampling
  xTaskCreatePinnedToCore(
    TaskAcquisitionCore0, 
    "AcquisitionTask", 
    4000, 
    NULL, 
    1, 
    NULL, 
    0 // Mengunci task secara absolut di Core 0
  );
}

void loop() {
  // Core 1 (default loop) dibiarkan kosong
  vTaskDelete(NULL);
}

void TaskAcquisitionCore0(void *pvParameters) {
  sensors_event_t humidity, temp;
  
  for (;;) {
    // Pengambilan sampel setiap 2 detik secara presisi
    vTaskDelay(2000 / portTICK_PERIOD_MS); 
    
    aht.getEvent(&humidity, &temp);

    float ema_temp, ema_humi;

    ema_temp = tempFilter.update(temp.temperature);
    ema_humi = humiFilter.update(humidity.relative_humidity);

    // Cetak langsung dalam format string yang dipisahkan koma (CSV)
    Serial.printf("%lu,%.2f,%.2f,%.2f,%.2f\n", millis(), temp.temperature, humidity.relative_humidity, ema_temp, ema_humi);
  }
}