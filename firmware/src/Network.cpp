#include <WiFi.h>
#include <WiFiClientSecure.h>
#include "Network.h"
#include "Config.h"
#include "Globals.h"

WiFiClientSecure espClientSecure;
PubSubClient mqttClient(espClientSecure);

void initNetwork() {
  WiFi.begin(wifi_ssid, wifi_pass);
  espClientSecure.setInsecure();
  mqttClient.setServer(mqtt_server, mqtt_port);
  mqttClient.setCallback(mqttCallback);
}

bool sendChunkToMQTT(Chunk &chunk) {
  uint8_t* binaryPayload = (uint8_t*)&chunk;
  size_t payloadLength = sizeof(Chunk);
  
  snprintf(dataTopic, sizeof(dataTopic), "devices/%d/data", chunk.nodeID);
  return mqttClient.publish(dataTopic, binaryPayload, payloadLength);
}

void mqttCallback(char* topic, byte* payload, unsigned int length) {
  Serial.printf("[MQTT DEBUG] Menerima pesan di topik: %s\n", topic);
  if (String(topic) == assignmentTopic && !isRegistered) {
    char idStr[length + 1];
    memcpy(idStr, payload, length);
    idStr[length] = '\0';
    
    int assignedID = atoi(idStr);
    if (assignedID > 0 && assignedID <= 255) {
      myNodeID = (uint8_t)assignedID;
      prefs.putUChar("node_id", myNodeID);
      isRegistered = true;
      Serial.printf("[MQTT] BERHASIL! Server memberikan Node ID: %d. Tersimpan permanen.\n", myNodeID);
    }
  }
}
