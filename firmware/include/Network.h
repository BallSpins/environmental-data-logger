#pragma once
#include "Types.h"

void initNetwork();
bool sendChunkToMQTT(Chunk &chunk);
void mqttCallback(char* topic, byte* payload, unsigned int length);
