#pragma once
#include <Arduino.h>

struct Data {
  int16_t humi;
  int16_t temp;
};

struct Chunk {
  uint32_t timestamp;
  uint8_t nodeID;
  uint8_t padding[3];
  Data data[10];
};
