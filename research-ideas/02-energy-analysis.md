# Energy Consumption Analysis of an ESP32-Based Environmental Monitoring Node

## Background

Power consumption is one of the primary design constraints for autonomous environmental monitoring systems.

Although many studies report average current consumption, relatively few isolate the energy contribution of each subsystem, including sensing, wireless communication, external flash storage, OLED display, and deterministic sampling.

A subsystem-level energy analysis provides quantitative information regarding the dominant energy consumers and potential optimization opportunities.

---

## Research Question

Which subsystem contributes the largest proportion of total energy consumption in a low-cost environmental monitoring node?

---

## Objectives

- Quantify energy consumption of each hardware subsystem.
- Compare operating modes under different configurations.
- Estimate battery lifetime under representative workloads.
- Identify dominant energy consumers.

---

## Hardware

- ESP32
- AHT10
- OLED Display
- W25Q32
- DS3231
- INA219 or USB Power Meter

---

## Independent Variables

- WiFi enabled/disabled
- OLED enabled/disabled
- Flash logging enabled/disabled
- Sampling interval
- MQTT transmission interval

---

## Dependent Variables

- Current consumption
- Average power
- Energy per sample
- Battery lifetime estimation

---

## Evaluation Metrics

### Average Current

Mean current during steady-state operation.

Unit:
- mA

---

### Average Power

Voltage × Average Current

Unit:
- W

---

### Energy per Sample

Total Energy / Number of Samples

Unit:
- J/sample

---

### Battery Lifetime

Battery Capacity / Average Current

Unit:
- hours

---

## Experimental Procedure

1. Measure baseline current with sensing only.
2. Enable OLED.
3. Enable WiFi.
4. Enable MQTT transmission.
5. Enable Store-and-Forward.
6. Enable all subsystems simultaneously.
7. Repeat measurements several times.

---

## Experimental Modes

| Mode | Active Components |
|------|-------------------|
| M1 | Sensor Only |
| M2 | Sensor + OLED |
| M3 | Sensor + WiFi |
| M4 | Sensor + Flash |
| M5 | Full System |

---

## Expected Analysis

- Current consumption comparison
- Energy contribution of each subsystem
- Battery lifetime estimation
- Dominant energy consumers
- Potential optimization strategies