# Firmware Subsystem

## Overview
The firmware subsystem is responsible for environmental data acquisition, noise mitigation, and telemetry transmission. It operates independently of the backend pipeline.

## Core Responsibilities
*   **Data Acquisition:** Reads raw environmental metrics from hardware sensors.
*   **Signal Processing (EMA Filter):** Applies an Exponential Moving Average (EMA) algorithm to raw sensor outputs to mitigate transient hardware noise and environmental fluctuations.
*   **Serialization:** Packs processed telemetry into binary (or JSON) payloads to minimize transmission overhead.
*   **Transmission:** Publishes the structured payloads to the backend broker via MQTT.
*   **Fault Tolerance (Store-and-Forward):** Implements a local FIFO buffer using non-volatile storage. Telemetry chunks are written to local storage (SD Card/Flash Memory) during network disruptions and sequentially transmitted once MQTT connectivity is restored, preventing data loss.

## Execution Environments
*   **Emulated (Current):** Utilizes Wokwi simulation with virtual SD Card mapping to validate logic, serialization, local queueing, and communication flow without physical hardware dependencies.
*   **Physical (Planned):** Target deployment on ESP32 hardware utilizing an onboard SPI Flash Memory (e.g., W25Q32) via LittleFS for robust physical sensor integration and non-volatile queue persistence.