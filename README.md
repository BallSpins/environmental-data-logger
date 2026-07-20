# Environmental Data Logger

## Project Overview
Environmental Data Logger is an IoT data acquisition system designed to collect, process, and store environmental telemetry. The architecture separates data acquisition at the edge from data persistence in the backend. 

To improve data validity at the hardware level, the system implements Exponential Moving Average (EMA) filtering on raw sensor data to attenuate physical noise before transmission.

## System Architecture
The pipeline utilizes a decoupled producer-consumer model.

*   **Edge Node:** Acquires sensor data, applies EMA filtering, and transmits payloads via MQTT.
*   **Ingestion:** Receives MQTT messages and buffers them immediately into a Redis FIFO queue.
*   **Processing:** Background workers consume the queue, deserialize the data, and execute batch inserts into a MySQL database.

## Repository Structure
*   `backend/`: Server-side services, message queue processing, and benchmarking utilities.
*   `firmware/`: Embedded application for data acquisition and EMA signal processing.
*   `docs/`: Technical documentation and benchmark metrics.

## Hardware Roadmap
*   **Current State:** Software-based device emulation via Wokwi for functional verification.
*   **Physical Deployment (Planned):** ESP32 deployment integrating physical environmental sensors and on-board hardware components.

## License
Refer to the [`LICENSE`](./LICENSE) file for the complete licensing terms.