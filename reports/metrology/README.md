# Metrology Report

## Overview

This report investigates the measurement characteristics of a low-cost environmental sensing system using an ESP32-based acquisition platform.

The primary objective is to evaluate measurement quality through deterministic data acquisition, signal conditioning, and quantitative noise analysis rather than proposing a new filtering algorithm.

## Research Scope

The report includes investigations on:

* Deterministic sensor acquisition using hardware timers.
* Environmental noise characterization.
* Exponential Moving Average (EMA) filtering.
* Mathematical optimization of the EMA smoothing factor.
* Comparison between theoretical ADC resolution and measured sensor noise.
* Signal stability and filtering performance.
* Experimental validation using repeated measurements.

## Hardware Platform

* ESP32
* AHT10 Temperature and Humidity Sensor

Additional hardware components will be evaluated in separate reports.

## Directory Structure

* `analysis/`
  Supporting calculations, derivations, and supplementary analyses.

* `assets/`
  Figures, plots, photographs, and experimental illustrations.

* `chapters/`
  Individual LaTeX chapters composing the complete report.

* `references/`
  Bibliography database and citation resources.

* `result/`
  Generated figures, processed datasets, and compiled report outputs.

## Related Reports

This report is part of a larger research series investigating different aspects of the Environmental Data Logger platform.

Planned reports include:

* Fault-Tolerant Store-and-Forward Architecture
* Energy Consumption Analysis
* End-to-End Latency Analysis

Each report addresses a distinct research question while sharing the same experimental platform.