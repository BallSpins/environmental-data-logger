# End-to-End Latency Analysis of an Environmental Monitoring Architecture

## Background

Most environmental monitoring studies focus primarily on measurement accuracy while providing limited discussion regarding communication latency.

For real-time monitoring systems, the elapsed time between sensor acquisition and persistent database storage is equally important.

A complete latency analysis enables identification of communication bottlenecks and provides quantitative evidence regarding overall system responsiveness.

---

## Research Question

What is the end-to-end latency of an ESP32-based environmental monitoring architecture under different communication loads?

---

## Objectives

- Measure complete system latency.
- Identify latency contribution of each processing stage.
- Determine the dominant communication bottleneck.
- Evaluate latency stability under different workloads.

---

## Hardware

- ESP32
- AHT10
- DS3231
- MQTT Broker
- Backend Server
- Database

---

## Latency Definition

Latency measurement begins immediately after sensor acquisition is completed and ends when the measurement has been successfully committed into the database.

---

## Latency Components

Total Latency consists of:

- Sensor Acquisition
- Signal Processing
- Queue Waiting
- Network Transmission
- MQTT Processing
- Backend Processing
- Database Commit

---

## Independent Variables

- MQTT QoS
- Payload size
- Sampling interval
- Batch size
- Redis cache enabled/disabled

---

## Dependent Variables

- Total latency
- Network latency
- Backend latency
- Database latency
- Latency jitter

---

## Evaluation Metrics

### Mean Latency

Average end-to-end delay.

---

### Median Latency

Robust central latency estimate.

---

### 95th Percentile Latency

Latency experienced by 95% of transmitted samples.

---

### Maximum Latency

Worst-case delay observed during the experiment.

---

### Throughput

Successfully processed samples per second.

---

## Experimental Procedure

1. Generate deterministic sensor measurements.
2. Timestamp each measurement immediately after acquisition.
3. Publish data through MQTT.
4. Record timestamp when data reaches the backend.
5. Record timestamp after database commit.
6. Compute latency for every sample.
7. Repeat under different workloads.

---

## Experimental Scenarios

| Scenario | Description |
|----------|-------------|
| S1 | Normal operation |
| S2 | Large payload |
| S3 | High sampling frequency |
| S4 | Batch transmission |
| S5 | Heavy backend workload |

---

## Expected Analysis

- End-to-end latency distribution
- Latency breakdown per processing stage
- Throughput comparison
- Latency jitter
- Identification of dominant bottlenecks