# Backend Subsystem

## Overview
The backend subsystem handles message ingestion, buffering, and database persistence. It isolates database latency from message reception throughput using an asynchronous worker pool.

## Pipeline Components
*   **Transport Layer (MQTT):** Serves strictly as the transport protocol for incoming telemetry.
*   **Buffer Layer (Redis):** Acts as an intermediate queue to absorb traffic bursts and decouple producers from consumers.
*   **Processing Layer (Go Workers):** Asynchronously fetches messages, validates payloads, and parses binary/JSON formats.
*   **Persistence Layer (MySQL):** Accumulates processed records and executes batched database insertions to maximize write throughput.

## Performance Benchmarking
This directory includes utilities to measure architectural efficiency. Current evaluations include:
*   Payload serialization latency and memory allocation (JSON vs. Binary).
*   Database batch insertion throughput.
*   Redis buffering behavior under stress conditions.

Benchmark results are documented in [`docs/metrics/`](../docs/metrics).

## Execution
1. Configure environment variables for MQTT, Redis, and MySQL.
2. Start infrastructure services.
3. Execute the worker pool application.