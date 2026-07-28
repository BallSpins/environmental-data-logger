import os
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from scipy.stats import norm
from scipy.stats import linregress

def analyze_sensor_data(file_path, target_col, output_dir, unit_label, col_name_nice):
    # 1. Pembersihan Data (Data Cleaning)
    header_idx = 0
    with open(file_path, 'r') as f:
        for i, line in enumerate(f):
            if 'timestamp_ms,temperature_c,humidity_rh' in line:
                header_idx = i
                break

    # Memuat DataFrame dan konversi ke numerik
    df = pd.read_csv(file_path, skiprows=header_idx)
    df = df.apply(pd.to_numeric, errors='coerce').dropna()

    # 2. Pemotongan Awal (Trimming)
    # Membuang 30 menit pertama (timestamp_ms < 1800000)
    df_trimmed = df[df['timestamp_ms'] >= 1800000].copy()
    df_trimmed['time_min'] = df_trimmed['timestamp_ms'] / 60000

    # 3. Kompensasi Drift (Linear Detrending)
    slope, intercept, r_value, p_value, std_err = linregress(df_trimmed['time_min'], df_trimmed[target_col])
    df_trimmed['trend'] = intercept + slope * df_trimmed['time_min']
    df_trimmed['detrended'] = df_trimmed[target_col] - df_trimmed['trend']
    r2 = r_value**2

    # 4. Kalkulasi Parameter Metrologi Dasar
    mean_initial = df_trimmed[target_col].mean()
    sigma_detrended = df_trimmed['detrended'].std()

    print(f"[{col_name_nice}] Average (μ): {mean_initial:.4f} {unit_label}")
    print(f"[{col_name_nice}] Slope (Trend): {slope:.6f} {unit_label}/minute")
    print(f"[{col_name_nice}] R-Squared (R2): {r2:.4f}")
    print(f"[{col_name_nice}] Standard Deviation After Detrending (σ): {sigma_detrended:.4f} {unit_label}")

    # Membuat folder output jika belum ada
    os.makedirs(output_dir, exist_ok=True)

    # 5. Visualisasi Time-Series
    plt.figure(figsize=(12, 8))

    plt.subplot(2, 1, 1)
    plt.plot(df_trimmed['time_min'], df_trimmed[target_col], label=f'Raw {col_name_nice} Data', color='blue', alpha=0.6)
    plt.plot(df_trimmed['time_min'], df_trimmed['trend'], label=f'Linear Trend (Slope: {slope:.4f} {unit_label}/min)', color='red', linewidth=2)
    plt.title(f'{col_name_nice} vs. Time (After 30 Minutes)')
    plt.ylabel(f'{col_name_nice} ({unit_label})')
    plt.grid(True, linestyle='--', alpha=0.6)
    plt.legend()

    plt.subplot(2, 1, 2)
    plt.plot(df_trimmed['time_min'], df_trimmed['detrended'], label='Detrended Signal', color='green', alpha=0.7)
    plt.axhline(0, color='red', linestyle='--', linewidth=1.5)
    plt.title('Pure Noise Signal (After Detrending)')
    plt.xlabel('Time (Minutes)')
    plt.ylabel(f'{col_name_nice} Deviation ({unit_label})')
    plt.grid(True, linestyle='--', alpha=0.6)
    plt.legend()

    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'timeseries_plot.png'), dpi=300)
    plt.close()

    # 6. Visualisasi Distribusi Noise
    plt.figure(figsize=(8, 6))
    plt.hist(df_trimmed['detrended'], bins=50, density=True, alpha=0.6, color='gray', label='Detrended Data Histogram')

    # Distribusi Normal Ideal
    mu_ideal = 0
    x = np.linspace(df_trimmed['detrended'].min(), df_trimmed['detrended'].max(), 100)
    pdf = norm.pdf(x, mu_ideal, sigma_detrended)
    plt.plot(x, pdf, 'r-', linewidth=2, label=f'Normal Distribution ($\mu=0$, $\sigma={sigma_detrended:.4f}$)')

    plt.title(f'AHT10 Sensor Noise Distribution After Detrending ({col_name_nice})')
    plt.xlabel(f'{col_name_nice} Deviation ({unit_label})')
    plt.ylabel('Probability Density')
    plt.grid(True, linestyle='--', alpha=0.6)
    plt.legend()
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'histogram_plot.png'), dpi=300)
    plt.close()

file_path = 'dataset_aht10_steady.csv'

# Analisis Temperatur (Menyimpan ke assets/T)
analyze_sensor_data(
    file_path=file_path,
    target_col='temperature_c',
    output_dir='./../assets/T',
    unit_label='°C',
    col_name_nice='Temperature'
)

# Analisis Kelembapan (Menyimpan ke assets/RH)
analyze_sensor_data(
    file_path=file_path,
    target_col='humidity_rh',
    output_dir='./../assets/RH',
    unit_label='%RH',
    col_name_nice='Humidity'
)