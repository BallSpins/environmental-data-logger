#pragma once
#include <Arduino.h>
#include <WiFi.h>
#include <WiFiClientSecure.h>
#include <PubSubClient.h>
#include <Preferences.h>
#include "Types.h"
#include "Filter.h"

extern QueueHandle_t dataQueue;
extern Preferences prefs;
extern WiFiClientSecure espClientSecure;
extern PubSubClient mqttClient;

extern EMAFilter tempFilter;
extern EMAFilter humiFilter;

extern uint8_t myNodeID;
extern bool isRegistered;
extern char assignmentTopic[36];
extern char dataTopic[36];
