import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

# Load Data
with open('dataset_aht10_transient.csv', 'r') as f:
    lines = f.readlines()
    
header_idx = -1
for i, line in enumerate(lines):
    if 'timestamp_ms' in line:
        header_idx = i
        break

df = pd.read_csv('dataset_aht10_transient.csv', skiprows=header_idx)
df = df.apply(pd.to_numeric, errors='coerce').dropna()
df['time_min'] = df['timestamp_ms'] / 60000.0

# EMA Function
def calculate_ema(data, alpha):
    ema = np.zeros(len(data))
    ema[0] = data.iloc[0]
    for i in range(1, len(data)):
        ema[i] = alpha * data.iloc[i] + (1 - alpha) * ema[i-1]
    return ema

# Apply EMAs dengan Alpha Optimal
alphas = [0.05, 0.1, 0.1494, 0.2, 0.4]
labels = [
    r'$\alpha = 0.05$ (Tmax $\approx$ 40s)',
    r'$\alpha = 0.1$ (Tmax $\approx$ 20s)',
    r'$\alpha = 0.1494$ (Tmax $\approx$ 13.4s) [OPTIMAL]',
    r'$\alpha = 0.2$ (Tmax $\approx$ 10s)',
    r'$\alpha = 0.4$ (Tmax $\approx$ 5s)'
]
colors = ['purple', 'blue', 'green', 'orange', 'red']

for alpha in alphas:
    df[f'ema_{alpha}'] = calculate_ema(df['temperature_c'], alpha)

# 1. Overall Plot
plt.figure(figsize=(14, 8))
plt.plot(df['time_min'], df['temperature_c'], label='Raw Data (Suhu Mentah)', color='black', alpha=0.3, linewidth=2)

for alpha, label, color in zip(alphas, labels, colors):
    # Menebalkan garis hijau (optimal) agar menonjol
    lw = 2.5 if alpha == 0.1494 else 1.5
    plt.plot(df['time_min'], df[f'ema_{alpha}'], label=label, color=color, linewidth=lw)

plt.axvline(x=3.0, color='gray', linestyle='--', label='Waktu Injeksi Botol Hangat (~3 Menit)')
plt.title('Uji Respons Transien (Phase Lag) Filter EMA AHT10 (Keseluruhan)')
plt.xlabel('Waktu (Menit)')
plt.ylabel('Suhu (°C)')
plt.legend()
plt.grid(True, linestyle='--', alpha=0.6)
plt.tight_layout()
plt.savefig('./../assets/transient_optimized_overall_plot.png', dpi=300)
plt.close()

# 2. Zoomed Plot (Phase Lag Highlight)
plt.figure(figsize=(14, 8))
plt.plot(df['time_min'], df['temperature_c'], label='Raw Data (Suhu Mentah)', color='black', alpha=0.4, marker='.', markersize=4)

for alpha, label, color in zip(alphas, labels, colors):
    # Menebalkan garis hijau (optimal) agar menonjol
    lw = 3.0 if alpha == 0.1494 else 1.5
    plt.plot(df['time_min'], df[f'ema_{alpha}'], label=label, color=color, linewidth=lw)

# Zoom around minute 2.5 to 5.5 to see the lag clearly
plt.xlim(2.5, 5.5)
plt.ylim(df['temperature_c'][df['time_min'] < 5.5].min() - 0.5, df['temperature_c'][(df['time_min'] > 3) & (df['time_min'] < 5.5)].max() + 0.5)

plt.axvline(x=3.0, color='gray', linestyle='--', label='Injeksi Panas')
plt.title('Zoom Transisi: Efek Phase Lag pada Loncatan Suhu')
plt.xlabel('Waktu (Menit)')
plt.ylabel('Suhu (°C)')
plt.legend(loc='lower right')
plt.grid(True, linestyle='--', alpha=0.6)
plt.tight_layout()
plt.savefig('./../assets/transient_optimized_zoomed_plot.png', dpi=300)
plt.close()