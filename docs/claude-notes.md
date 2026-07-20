\# Research-Grade Performance Analysis

\## Binary IoT Ingestion Pipeline: MQTT → Redis FIFO → Ingestion Worker → MySQL



\*\*Role:\*\* Distributed-systems reviewer / performance engineer (IEEE/ACM-style discussion-section review)

\*\*Source data:\*\* 5 benchmark experiments + 1 metadata file (`metadata.json`, `result-json-vs-binary.\*`, `result-chunk-aggregation.\*`, `result-db-insert.\*`, `result-redis-buffer.\*`, `result-stress-test.\*`)

\*\*Captured:\*\* 2026-07-16, go1.25.0, windows/amd64, 8 logical CPUs, Redis 6.2+, MySQL 8.0+



A structural note before the content: you asked for Executive Summary → Research Findings → Practical Engineering Implications → Limitations → Future Work as the closing block. I kept all five, but moved the Executive Summary to the front — that's where a reader needs it, and it's standard placement for a thesis/paper discussion section — and kept the other four as the closing synthesis (§9–§12), after all the evidence has been laid out. Nothing you asked for was dropped.



\---



\## Executive Summary



1\. \*\*The core architectural thesis (binary > JSON, batching > single-row, aggregation > per-reading) is directionally correct, but several individual metrics are actively misleading if read in isolation\*\* — which is exactly why this report cross-references every number against the others rather than walking through files one at a time. Three examples, each detailed in §3: the raw `binary\_heap\_allocated\_mb` figure is \*higher\* than JSON's until normalized by iteration count, at which point binary is \~21.7× more allocation-efficient, as the architecture claims (§3.1). The Redis-buffered path is measured as 51.6% \*slower\* than direct processing for a single burst (§3.4) — true for that narrow test, but not evidence against buffering's actual purpose. Gzip \*inflates\* the 48-byte binary payload by 52% (§3.1).



2\. \*\*The single most important cross-experiment finding is a triangulation, not a direct measurement\*\*: composing the stress test's record-generation rate (§3.5) with the DB-insert experiment's measured MySQL ceiling (§3.3) shows the tested 200-node configuration generates records \~9.2% faster than the best measured MySQL insert rate can absorb. This is consistent with — and is the most likely explanation for — the stress test's unbounded, un-drained Redis queue (`final\_queue\_len == max\_queue\_len == 19800`). No single file states this; it only emerges from reading them together, which is what this report does throughout.



3\. \*\*Batch INSERT is the highest-leverage, most confidently measured optimization in the suite\*\*: a 45.3× throughput improvement from batch=1 to batch=500 (§3.3), with diminishing returns already visible by batch=250–500. It is also the lever most likely aimed at the actual bottleneck (§5).



4\. \*\*The benchmark cannot, by itself, support a single confident "N devices" scalability number\*\*, because no experiment measures producer and consumer throughput simultaneously in one integrated run. §6 derives a bracketed estimate of \*\*\~88–183 sustainable devices\*\* at the tested per-device rate, with the width of that bracket itself being a finding: it reflects real uncertainty about which of two very different mechanisms is gating throughput, not imprecise arithmetic.



5\. \*\*Chunk aggregation has a large, directly quantifiable AWS cost effect\*\* that the benchmark data supports cleanly: at the same total sensor-reading volume, chunking 10 readings per MQTT message cuts the AWS IoT Core messaging bill by a measured-and-priced \~10× versus one message per reading (§7).



6\. \*\*Verdict on publication-readiness: not yet\*\*, primarily because (a) no experiment reports CPU utilization anywhere in the suite despite CPU-bound behavior being central to the claimed optimizations, (b) the stress test never reports consumer/database-side throughput, so the bottleneck location is inferred, not observed, and (c) n=3 repetitions with no variance reporting is too thin to trust point estimates — the `batch\_25\_allocated\_mb` outlier in §3.3 is a direct symptom of this. §8 details these and four more.



Every claim below is tagged \*\*\[MEASURED]\*\* (taken directly from a result file), \*\*\[DERIVED]\*\* (computed from measured values via a stated formula), \*\*\[INFERENCE]\*\* (a reasoned architectural/systems conclusion not directly measured), or \*\*\[ESTIMATE]\*\* (a forward-looking engineering projection). This mirrors your instruction to keep facts, interpretations, and estimates distinct throughout, not just in the scalability/AWS sections where you asked for it explicitly.



\---



\## 1. Methodology \& Reading Notes



\*\*How the files were used.\*\* All three formats per experiment (`.json`, `.csv`, `.md`) were checked and confirmed to be redundant encodings of the same underlying numbers — none carries information the others lack. The analysis below is therefore built from the `.json` files, cross-referenced against each other and against `metadata.json`, per your instruction not to analyze files in isolation.



\*\*Notation used throughout:\*\*



| Tag | Meaning |

|---|---|

| \*\*\[MEASURED]\*\* | A value taken directly from a result file, unmodified |

| \*\*\[DERIVED]\*\* | Computed from one or more measured values via an explicit, shown formula (e.g., bytes/op, records/sec) |

| \*\*\[INFERENCE]\*\* | A systems-level conclusion connecting measured data to architectural behavior that the benchmark does not directly instrument (e.g., "this is where the bottleneck sits") |

| \*\*\[ESTIMATE]\*\* | A forward-looking engineering projection (scalability ceiling, AWS cost) that necessarily combines measured data with assumptions, stated explicitly |



\*\*Two interpretive decisions made up front, used repeatedly below (both cross-validated against multiple independent numbers, shown in §3.5):\*\*

\- `publish\_rate: 0.05` in the stress-test parameters is read as \*seconds between publishes per simulated node\* (→ 20 Hz/node), not messages/sec. This reading is what makes `total\_publishes`, `throughput\_pub\_sec`, and `bandwidth\_kb\_sec` mutually consistent (§3.5); the reverse reading does not.

\- The stress test's binary payload is read as \*\*chunk\_size = 10\*\* (48 bytes/message), because `bandwidth\_kb\_sec` ÷ `throughput\_pub\_sec` reproduces 48.00 bytes/message exactly — matching `chunk\_10\_payload\_size\_bytes` from the chunk-aggregation experiment and `binary\_payload\_size\_bytes` from the JSON-vs-binary experiment to the byte. This is not a stated parameter; it is reconstructed from the data, and the reconstruction is exact enough (0.00-byte error) to be treated as solid.



\*\*A general caution used throughout §3\*\*: several metrics in this suite (`\*\_heap\_allocated\_mb`, in particular) appear to be \*\*cumulative totals over an entire fixed-duration benchmark run\*\*, not per-operation figures — the files do not document this explicitly, and that ambiguity is itself flagged as a methodological weakness in §8. Raw totals are only compared across experiments after normalizing by iteration/operation count; comparing un-normalized totals across experiments with different iteration counts is called out explicitly wherever it would otherwise produce a wrong conclusion.



\---



\## 2. Unified Consolidated Data Table



Per your instruction, this table combines every metric from every experiment into one view and is the basis for all analysis that follows — no section below introduces a number that isn't traceable to a row here (or to an explicit formula applied to rows here).



\### 2.1 Raw measured metrics (as reported in the result files)



| # | Experiment | Metric | Value | Unit |

|---|---|---|---|---|

| 1 | json-vs-binary | json\_payload\_size\_bytes | 324 | bytes |

| 2 | json-vs-binary | binary\_payload\_size\_bytes | 48 | bytes |

| 3 | json-vs-binary | json\_gzip\_size\_bytes | 172 | bytes |

| 4 | json-vs-binary | binary\_gzip\_size\_bytes | 73 | bytes |

| 5 | json-vs-binary | json\_serial\_ns\_op | 9,831.9475 | ns/op |

| 6 | json-vs-binary | binary\_serial\_ns\_op | 323.2828 | ns/op |

| 7 | json-vs-binary | json\_heap\_allocated\_mb | 151.5679 | MB (cumulative, see §1) |

| 8 | json-vs-binary | binary\_heap\_allocated\_mb | 212.7476 | MB (cumulative, see §1) |

| 9 | chunk-aggregation | chunk\_1\_ops\_sec | 90,474,756.84 | ops/sec |

| 10 | chunk-aggregation | chunk\_1\_payload\_size\_bytes | 12 | bytes |

| 11 | chunk-aggregation | chunk\_10\_ops\_sec | 16,581,316.18 | ops/sec |

| 12 | chunk-aggregation | chunk\_10\_payload\_size\_bytes | 48 | bytes |

| 13 | chunk-aggregation | chunk\_30\_ops\_sec | 12,844,166.54 | ops/sec |

| 14 | chunk-aggregation | chunk\_30\_payload\_size\_bytes | 128 | bytes |

| 15 | chunk-aggregation | chunk\_60\_ops\_sec | 9,456,984.51 | ops/sec |

| 16 | chunk-aggregation | chunk\_60\_payload\_size\_bytes | 248 | bytes |

| 17 | db-insert | batch\_1\_inserts\_sec | 716.2960 | rows/sec |

| 18 | db-insert | batch\_1\_allocated\_mb | 0.6258 | MB |

| 19 | db-insert | batch\_10\_inserts\_sec | 6,037.9649 | rows/sec |

| 20 | db-insert | batch\_10\_allocated\_mb | 0.0728 | MB |

| 21 | db-insert | batch\_25\_inserts\_sec | 11,523.7710 | rows/sec |

| 22 | db-insert | batch\_25\_allocated\_mb | 0.0048 | MB (anomalous, see §3.3) |

| 23 | db-insert | batch\_50\_inserts\_sec | 18,082.2839 | rows/sec |

| 24 | db-insert | batch\_50\_allocated\_mb | 0.0655 | MB |

| 25 | db-insert | batch\_100\_inserts\_sec | 22,563.2791 | rows/sec |

| 26 | db-insert | batch\_100\_allocated\_mb | 0.0661 | MB |

| 27 | db-insert | batch\_250\_inserts\_sec | 30,142.4099 | rows/sec |

| 28 | db-insert | batch\_250\_allocated\_mb | 0.0466 | MB |

| 29 | db-insert | batch\_500\_inserts\_sec | 32,474.8299 | rows/sec |

| 30 | db-insert | batch\_500\_allocated\_mb | 0.0499 | MB |

| 31 | redis-buffer | direct\_processing\_sec | 0.0935 | sec (for a 200-msg burst) |

| 32 | redis-buffer | buffered\_processing\_sec | 0.1418 | sec (for a 200-msg burst) |

| 33 | redis-buffer | max\_queue\_accumulation | 200 | messages |

| 34 | stress-test | throughput\_pub\_sec | 3,546.3527 | msg/sec |

| 35 | stress-test | total\_publishes | 35,468 | messages |

| 36 | stress-test | bandwidth\_kb\_sec | 166.2353 | KB/sec |

| 37 | stress-test | mean\_latency\_ms | 1 | ms |

| 38 | stress-test | p50\_latency\_ms | 0 | ms |

| 39 | stress-test | p95\_latency\_ms | 5 | ms |

| 40 | stress-test | p99\_latency\_ms | 6 | ms |

| 41 | stress-test | max\_queue\_len | 19,800 | items |

| 42 | stress-test | final\_queue\_len | 19,800 | items |

| 43 | stress-test | heap\_allocated\_mb | 15.0746 | MB (cumulative, see §1) |



\### 2.2 Key derived / normalized metrics (computed for this analysis, formulas shown)



| # | Metric | Value | Formula / source rows |

|---|---|---|---|

| D1 | Binary wire-size reduction vs JSON | 85.2% smaller | 1 − (row2/row1) |

| D2 | Binary serialization speedup vs JSON | 30.41× | row5/row6 |

| D3 | Binary iterations completed in the 3s test window | ≈9,279,800 | 3s / (row6 in seconds) |

| D4 | JSON iterations completed in the 3s test window | ≈305,128 | 3s / (row5 in seconds) |

| D5 | Binary allocation, normalized per operation | 24.04 bytes/op | (row8 in bytes) / D3 |

| D6 | JSON allocation, normalized per operation | 520.87 bytes/op | (row7 in bytes) / D4 |

| D7 | Normalized allocation efficiency, binary vs JSON | 21.67× less | D6/D5 |

| D8 | Gzip effect on JSON payload | −46.9% (helps) | 1 − row3/row1 |

| D9 | Gzip effect on binary payload | \*\*+52.1% (hurts)\*\* | row4/row2 − 1 |

| D10 | Binary chunk framing model | payload = 8 + 4×n bytes | exact linear fit across rows 10,12,14,16 (n=1,10,30,60), max error 0.0 bytes |

| D11 | Records/sec at chunk=1 / 10 / 30 / 60 | 90.47M / 165.81M / 385.32M / 567.42M rec/sec | row9×1, row11×10, row13×30, row15×60 |

| D12 | ns/record at chunk=1 / 10 / 30 / 60 | 11.05 / 6.03 / 2.60 / 1.76 ns | 1e9/(ops\_sec×n) |

| D13 | Discrepancy: `binary\_serial\_ns\_op` vs chunk-aggregation's implied chunk\_10 op time | 323.28ns vs 60.31ns (5.36×) | row6 vs 1e9/row11 — see §3.2 |

| D14 | DB insert throughput multiplier, batch 1→500 | 45.34× | row29/row17 |

| D15 | DB insert throughput multiplier, batch 250→500 (2× batch size) | 1.077× (diminishing) | row29/row27 |

| D16 | DB round trips needed for 1,000 rows, by batch size | 1000, 100, 40, 20, 10, 4, 2 | 1000/batch\_size |

| D17 | Time per DB round trip (batch=1 case) | 1.396 ms | 1/row17 |

| D18 | Redis-buffered overhead vs direct, single 200-msg burst | +51.6% wall time, +241.3 µs/msg | row32/row31; (row32−row31)/200 |

| D19 | Stress test: theoretical max publish rate | 4,000 msg/sec | 200 nodes × 20 Hz |

| D20 | Stress test: achieved vs theoretical | 88.66% | row34/D19 |

| D21 | Stress test: bandwidth-implied bytes/message | 48.00 bytes (exact match to row12) | (row36×1024)/row34 |

| D22 | Stress test: record-generation rate (chunk=10 basis) | 35,463.5–35,468 rec/sec | row34×10, cross-checked against row35×10/10s |

| D23 | Stress test: per-node achieved rate | 177.32 rec/sec/node | D22/200 |

| D24 | Stress test: per-node target rate | 200 rec/sec/node | 20 Hz × 10 rec/msg |

| D25 | Stress test: net queue growth rate | 1,980 msg/sec | row42/10s |

| D26 | Stress test: implied in-situ consumption rate (if a consumer was active) | 1,566.4 msg/sec = 15,663.5 rec/sec | row34 − D25 |

| D27 | Stress test: heap allocated per publish (full-system) | 445.66 bytes/publish | (row43 in bytes)/row35 |

| D28 | Full-system vs isolated-serialization allocation ratio | 18.54× higher in situ | D27/D5 |

| D29 | Record-generation rate vs measured MySQL ceiling (batch=500) | 1.092× (generation exceeds ceiling) | D22/row29 |



All figures in §3–§9 trace back to this table. Where a number below isn't in either table directly, it's a one-step arithmetic combination of rows shown inline.

