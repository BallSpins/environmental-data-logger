# Experimental Design and Research Proposal

## 1. Rationale for Deferring the Abstract
Writing an abstract prior to physical data collection violates fundamental scientific methodology. An abstract requires empirical results, precise metrological metrics (e.g., error rates, filter time delays), and validated conclusions. Because the physical hardware is not yet constructed, these metrics do not exist. Drafting an abstract at this stage would result in a fabricated hypothesis masquerading as an empirical conclusion.

## 2. Actionable LaTeX Drafting Plan
Until the physical experiments are executed, the LaTeX drafting process must be strictly limited to the **Research Proposal** or **Experimental Design** phases. 

Focus exclusively on writing the following sections:

### A. Introduction
*   **Context:** The necessity of representative and high-integrity environmental data in remote edge-computing monitoring.
*   **Problem:** Hardware-induced transient noise in low-cost sensors (e.g., AHT10) and the susceptibility of real-time telemetry to network-induced packet loss.
*   **Objective:** To design and methodologically validate a data acquisition architecture that ensures data integrity through on-board signal processing and non-volatile local buffering.

### B. Experimental Methodology
*   **Digital Signal Processing (DSP):** Document the mathematical model of the Exponential Moving Average (EMA) filter.
    *   Formula: $y_t = \alpha x_t + (1-\alpha)y_{t-1}$
    *   Define all variables ($y_t$, $x_t$, $\alpha$) and explain the theoretical impact of the smoothing factor on transient noise mitigation.
*   **Fault Tolerance Architecture:** Describe the logical flow of the Store-and-Forward mechanism using the local SD Card buffer.
    *   Detail the system transition states: Network failure triggers local write operations to non-volatile memory; network recovery triggers FIFO queue flushing to the MQTT broker.

## 3. Pending Sections (Placeholder Only)
Leave the following sections entirely blank in the `main.tex` file until physical testing is complete:
*   **Abstract:** Deferred until all data is collected and statistically analyzed.
*   **Results & Analysis:** Deferred until the generation of overlay graphs (raw vs. filtered signal data) and disruption recovery matrices.
*   **Conclusion:** Deferred until empirical validation of the system's data integrity is achieved.