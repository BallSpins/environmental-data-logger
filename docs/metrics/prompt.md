You are acting as a senior Go performance engineer.

I need you to implement a complete benchmarking and stress-testing suite for this project.

Before writing any code, first understand the existing backend architecture, especially the ingestion pipeline:

IoT Node
→ MQTT Broker
→ MQTT Handler
→ Redis FIFO Queue
→ Ingestion Worker
→ Parser
→ Batch Repository
→ MySQL

The current backend is heavily optimized around:

- Binary payload (48-byte chunk)
- Zero heap allocation during steady-state batches
- sync.Pool
- unsafe.String
- unsafe.Slice
- Batch INSERT
- Redis FIFO buffering

The purpose of this benchmarking suite is NOT to benchmark AWS.

The purpose is to collect metrics from the current local architecture which will later be used to estimate AWS operational cost.

Read the existing documentation first before implementing anything.

Relevant documentation:

./docs/backend.md
./docs/performance.md
./docs/metrics/plan.md

Follow the benchmarking plan defined inside:

./docs/metrics/plan.md

The benchmarking suite must cover these experiments:

1. JSON vs Binary
--------------------------------

Measure:

- payload size
- serialization time
- deserialization time
- CPU usage
- bandwidth
- throughput
- latency

Implement both:

- isolated benchmark (application only)
- backend pipeline benchmark
- full end-to-end benchmark

2. Chunk Aggregation
--------------------------------

Compare:

Chunk Size:

- 1
- 10
- 30
- 60

Measure:

- MQTT publish count
- payload size
- throughput
- end-to-end latency
- estimated AWS IoT message count

3. Single INSERT vs Batch INSERT
--------------------------------

Measure:

- inserts/sec
- commit latency
- CPU usage
- throughput
- rows/sec

Benchmark using identical generated datasets.

4. Redis Buffer vs Direct MySQL
--------------------------------

Compare:

MQTT
→ MySQL

vs

MQTT
→ Redis
→ Worker
→ MySQL

Measure:

- throughput
- burst handling
- queue growth
- processing latency
- dropped requests (if any)

5. End-to-End Scalability Test
--------------------------------

Create a stress testing tool capable of simulating many virtual IoT devices.

The simulator must support configurable virtual nodes:

100
500
1000
5000
10000
20000

Each virtual node generates exactly one chunk every 20 seconds.

The simulator must support:

- Binary payload
- JSON payload

Measure:

- total throughput
- CPU usage
- memory usage
- queue length
- database inserts/sec
- end-to-end latency
- publish/sec

--------------------------------

Implementation Requirements

DO NOT place these tools inside internal packages.

Create executable commands under:

./backend/cmd/

Suggested layout:

cmd/
    benchmark-json/
    benchmark-binary/
    benchmark-chunk/
    benchmark-db/
    benchmark-redis/
    stress-shadow-device/
    metrics-export/

Each command should be runnable independently.

Every command should expose flags such as:

--nodes
--duration
--payload=json|binary
--chunk-size
--batch-size
--output
--mqtt
--redis
--mysql

Use Cobra only if the project already uses it.
Otherwise keep standard Go flags.

--------------------------------

Output

Every benchmark must export:

CSV

JSON

Summary table

All metrics should include timestamps.

CSV format should be easy to import into Excel.

--------------------------------

Documentation

After implementing each benchmark:

Create documentation under:

./docs/metrics/

Suggested files:

overview.md

json-vs-binary.md

chunk-size.md

batch-insert.md

redis-buffer.md

stress-test.md

shadow-device.md

Each document should explain:

- benchmark objective
- architecture under test
- benchmark methodology
- configurable parameters
- measured metrics
- expected output
- how to execute
- example command
- sample output
- interpretation guide

--------------------------------

Code Quality

- idiomatic Go
- reusable components
- avoid duplicated benchmark logic
- deterministic benchmarks
- configurable workloads
- no magic numbers
- clear package separation

Do NOT modify production logic unless absolutely necessary.

If additional reusable utilities are needed, place them under an appropriate internal package.

Before coding, first generate an implementation plan describing:

- new directories
- new executables
- shared utilities
- benchmark flow
- documentation structure

Only after the plan is approved should implementation begin.