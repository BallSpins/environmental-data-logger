# Research Reports

## Overview

This directory contains technical reports documenting the experimental evaluation of individual research topics within the Environmental Data Logger project.

Each report investigates a specific engineering or scientific problem independently, including its methodology, experimental setup, analysis, and conclusions.

Although all reports are based on the same hardware and software platform, each focuses on a different research objective.

## Current Reports

| Report | Status | Focus |
|---------|--------|-------|
| Metrology | In Progress | Measurement quality, signal conditioning, and deterministic data acquisition |
| Fault Tolerance | Planned | Store-and-Forward architecture and communication reliability |
| Energy Analysis | Planned | Power consumption characterization |
| End-to-End Latency | Planned | System-wide latency evaluation |

## Report Structure

Each report follows a similar organization:

```
report-name/
├── analysis/
├── assets/
├── chapters/
├── references/
└── result/
```

where:

* `analysis/` contains supporting calculations and exploratory analyses.
* `assets/` stores figures and graphical resources.
* `chapters/` contains the LaTeX source for each report section.
* `references/` contains bibliography files.
* `result/` stores generated outputs and compiled artifacts.

## Scope

The reports are intended to document reproducible experiments rather than software implementation details. Source code is maintained separately within the firmware and backend subsystems.