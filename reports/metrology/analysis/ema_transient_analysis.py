import os
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

# EMA Function
def calculate_ema(data, alpha):
    ema = np.zeros(len(data))
    ema[0] = data.iloc[0]
    for i in range(1, len(data)):
        ema[i] = alpha * data.iloc[i] + (1 - alpha) * ema[i-1]
    return ema

def analyze_transient(file_path, target_col, output_dir, unit_label, col_name_nice):
    header_idx = 0
    with open(file_path, 'r') as f:
        for i, line in enumerate(f):
            if 'timestamp_ms,temperature_c,humidity_rh' in line:
                header_idx = i
                break

    df = pd.read_csv(file_path, skiprows=header_idx)
    df = df.apply(pd.to_numeric, errors='coerce').dropna()
    df['time_min'] = df['timestamp_ms'] / 60000.0

    # Apply EMAs
    alphas = [0.05, 0.0666, 0.1, 0.2, 0.4]
    labels = [
        r'$\alpha = 0.05$ (Tmax $\approx$ 40s)',
        r'$\alpha = 0.0666$ (Tmax $\approx$ 30s)',
        r'$\alpha = 0.1$ (Tmax $\approx$ 20s)',
        r'$\alpha = 0.2$ (Tmax $\approx$ 10s)',
        r'$\alpha = 0.4$ (Tmax $\approx$ 5s)'
    ]
    colors = ['purple', 'blue', 'green', 'orange', 'red']

    for alpha in alphas:
        df[f'ema_{alpha}'] = calculate_ema(df[target_col], alpha)

    # 1. Overall Plot
    plt.figure(figsize=(14, 8))
    plt.plot(df['time_min'], df[target_col], label='Raw Data', color='black', alpha=0.3, linewidth=2)

    for alpha, label, color in zip(alphas, labels, colors):
        plt.plot(df['time_min'], df[f'ema_{alpha}'], label=label, color=color, linewidth=1.5)

    plt.axvline(x=3.0, color='gray', linestyle='--', label='Warm Bottle Injection Time (~3 Minutes)')
    plt.title(f'AHT10 EMA Filter Transient Response (Phase Lag) - {col_name_nice}')
    plt.xlabel('Time (Minutes)')
    plt.ylabel(f'{col_name_nice} ({unit_label})')
    plt.legend()
    plt.grid(True, linestyle='--', alpha=0.6)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'transient_overall_plot.png'), dpi=300)
    plt.close()

    # 2. Zoomed Plot (Phase Lag Highlight)
    plt.figure(figsize=(14, 8))
    plt.plot(df['time_min'], df[target_col], label='Raw Data', color='black', alpha=0.4, marker='.', markersize=4)
    
    for alpha, label, color in zip(alphas, labels, colors):
        plt.plot(df['time_min'], df[f'ema_{alpha}'], label=label, color=color, linewidth=2)

    # Zoom around minute 2.5 to 5.5 to see the lag clearly
    plt.xlim(2.5, 5.5)
    plt.ylim(df[target_col][df['time_min'] < 5.5].min() - 0.5, df[target_col][(df['time_min'] > 3) & (df['time_min'] < 5.5)].max() + 0.5)

    plt.axvline(x=3.0, color='gray', linestyle='--', label='Heat Injection')
    plt.title(f'Transition Zoom: Phase Lag Effect on Step Response ({col_name_nice})')
    plt.xlabel('Time (Minutes)')
    plt.ylabel(f'{col_name_nice} ({unit_label})')
    plt.legend(loc='lower right')
    plt.grid(True, linestyle='--', alpha=0.6)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'transient_zoomed_plot.png'), dpi=300)
    plt.close()

file_path = 'dataset_aht10_transient.csv'

analyze_transient(
    file_path=file_path,
    target_col='temperature_c',
    output_dir='./../assets/T',
    unit_label='°C',
    col_name_nice='Temperature'
)

analyze_transient(
    file_path=file_path,
    target_col='humidity_rh',
    output_dir='./../assets/RH',
    unit_label='%RH',
    col_name_nice='Humidity'
)