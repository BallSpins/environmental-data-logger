#include <SD.h>
#include "Storage.h"
#include "Network.h"

void writeChunkToSD(Chunk &chunk) {
  File dFile = SD.open("/datalog.bin", FILE_APPEND);
  if (dFile) {
    dFile.write((uint8_t*)&chunk, sizeof(Chunk));
    dFile.close();
    Serial.println("[Storage] Append chunk ke SD Card sukses.");
  } else {
    Serial.println("[Storage] Gagal membuka file untuk menyimpan.");
  }
}

void processSDCardQueue() {
  File dataFile = SD.open("/datalog.bin", FILE_READ);
  if (!dataFile) return;

  Serial.println("[Storage] Menguras data antrean dari SD Card...");
  Chunk tempChunk;
  bool allSentSuccessfully = true;

  while (dataFile.read((uint8_t*)&tempChunk, sizeof(Chunk)) == sizeof(Chunk)) {
    if (sendChunkToMQTT(tempChunk)) {
      Serial.println("[Storage] Satu data antrean sukses dikirim ke MQTT.");
    } else {
      Serial.println("[Storage] Pengiriman antrean terhenti (MQTT putus lagi).");
      allSentSuccessfully = false;
      break; 
    }
  }
  dataFile.close();

  if (allSentSuccessfully) {
    SD.remove("/datalog.bin");
    Serial.println("[Storage] Seluruh antrean bersih. File datalog.bin dihapus.");
  }
}