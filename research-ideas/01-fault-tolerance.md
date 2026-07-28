# Fault-Tolerant Store-and-Forward Architecture for Environmental Monitoring

## Background

Environmental monitoring systems frequently operate in environments with unstable network connectivity. During communication outages, conventional IoT nodes often discard measurements or temporarily suspend data acquisition until connectivity is restored, resulting in permanent data gaps.

To mitigate this issue, a Store-and-Forward architecture can be implemented using external SPI Flash memory (W25Q32). During network outages, measurements are temporarily stored in external flash memory and automatically transmitted once the MQTT connection is re-established.

The primary objective is not to increase transmission speed, but to preserve data integrity and maximize recovery after prolonged communication failures.

---

## Research Question

Can a Store-and-Forward architecture based on ESP32 and external SPI Flash maintain complete data recovery during MQTT communication failures while preserving deterministic data acquisition?

---

## Objectives

- Design a fault-tolerant data acquisition architecture.
- Evaluate data recovery capability during network outages.
- Measure system recovery performance after connectivity restoration.
- Analyze flash utilization under prolonged communication failures.

---

## Hardware

- ESP32
- AHT10
- W25Q32 SPI Flash
- DS3231 RTC
- WiFi Router
- MQTT Broker

---

## Independent Variables

- Network outage duration
- Sampling interval
- MQTT QoS
- Flash utilization
- Payload size

---

## Dependent Variables

- Recovery Rate
- Data Loss
- Flush Throughput
- Recovery Delay
- Flash Occupancy

---

## Evaluation Metrics

### Recovery Rate

RR = Recovered Samples / Generated Samples

Target:
- 100%

---

### Data Loss

Generated Samples − Recovered Samples

Target:
- 0 samples

---

### Flush Throughput

Recovered Samples / Flush Time

Unit:
- samples/sec

---

### Recovery Delay

Elapsed time from network restoration until all buffered data has been transmitted.

---

### Flash Occupancy

Percentage of flash storage utilized during communication outages.

---

## Experimental Procedure

1. Configure deterministic sampling using hardware timer.
2. Disconnect the network for a predefined duration.
3. Continue data acquisition while storing measurements into W25Q32.
4. Restore network connectivity.
5. Automatically flush buffered records to the MQTT broker.
6. Compare generated and recovered datasets.
7. Repeat for multiple outage durations.

---

## Experimental Scenarios

| Test | Outage Duration |
|------|-----------------|
| T1 | 1 minute |
| T2 | 5 minutes |
| T3 | 15 minutes |
| T4 | 30 minutes |
| T5 | 60 minutes |

---

## Expected Analysis

- Recovery Rate versus outage duration
- Flush throughput
- Recovery latency
- Flash occupancy trend
- Queue growth during outage
- Data integrity verification