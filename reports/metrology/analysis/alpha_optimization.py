import os
import pandas as pd
import numpy as np
from scipy.stats import linregress
import matplotlib.pyplot as plt

# ==========================================
# 1. PARAMETER OPTIMASI
# ==========================================
W_NOISE = 0.5  # Bobot kepentingan meredam derau (50%)
W_LAG = 0.5    # Bobot kepentingan kecepatan respons (50%)
# W_NOISE = 0.3  # Bobot kepentingan meredam derau (30%)
# W_LAG = 0.7    # Bobot kepentingan kecepatan respons (70%)

# ==========================================
# 2. FUNGSI EKSEKUSI EMA
# ==========================================
def calculate_ema(data_series, alpha):
    ema = np.zeros(len(data_series))
    ema[0] = data_series.iloc[0]
    for i in range(1, len(data_series)):
        ema[i] = alpha * data_series.iloc[i] + (1 - alpha) * ema[i-1]
    return ema

def optimize_alpha(file_path_steady, file_path_transient, target_col, output_dir, col_name_nice):
    # ==========================================
    # 3. PREPARASI DATA KONDISI TUNAK (PROTOKOL 1)
    # ==========================================
    print("Memuat dataset kondisi tunak...")
    with open(file_path_steady, 'r') as f:
        lines_steady = f.readlines()
    header_idx_steady = next(i for i, line in enumerate(lines_steady) if 'timestamp_ms' in line)

    df_steady = pd.read_csv(file_path_steady, skiprows=header_idx_steady)
    df_steady = df_steady.apply(pd.to_numeric, errors='coerce').dropna()

    # Isolasi 30 menit ke atas dan lakukan linear detrending (menghapus thermal drift)
    df_steady = df_steady[df_steady['timestamp_ms'] >= 1800000].copy()
    slope, intercept, _, _, _ = linregress(df_steady['timestamp_ms'], df_steady[target_col])
    df_steady['detrended_noise'] = df_steady[target_col] - (slope * df_steady['timestamp_ms'] + intercept)

    # ==========================================
    # 4. PREPARASI DATA TRANSIEN (PROTOKOL 2)
    # ==========================================
    print("Memuat dataset transien...")
    with open(file_path_transient, 'r') as f:
        lines_trans = f.readlines()
    header_idx_trans = next(i for i, line in enumerate(lines_trans) if 'timestamp_ms' in line)

    df_trans = pd.read_csv(file_path_transient, skiprows=header_idx_trans)
    df_trans = df_trans.apply(pd.to_numeric, errors='coerce').dropna()
    df_trans['time_min'] = df_trans['timestamp_ms'] / 60000.0

    # Masking untuk area kalkulasi RMSE (menit 3.0 hingga 6.0)
    window_mask = (df_trans['time_min'] >= 3.0) & (df_trans['time_min'] <= 6.0)

    # ==========================================
    # 5. KALKULUS ITERATIF (COST FUNCTION)
    # ==========================================
    print("Mengeksekusi iterasi limit alpha...")
    alphas = np.linspace(0.01, 0.99, 500) # Evaluasi 500 titik dari 0.01 hingga 0.99
    cost_noise = []
    cost_lag = []

    for alpha in alphas:
        # A. Evaluasi Penalti Derau
        ema_noise = calculate_ema(df_steady['detrended_noise'], alpha)
        std_noise = np.std(ema_noise)
        cost_noise.append(std_noise)

        # B. Evaluasi Penalti Phase Lag
        ema_trans_full = calculate_ema(df_trans[target_col], alpha)
        raw_window = df_trans.loc[window_mask, target_col]
        ema_window = ema_trans_full[window_mask]
        rmse = np.sqrt(np.mean((raw_window - ema_window)**2))
        cost_lag.append(rmse)

    cost_noise = np.array(cost_noise)
    cost_lag = np.array(cost_lag)

    # ==========================================
    # 6. NORMALISASI MIN-MAX & KALKULASI TITIK MINIMUM
    # ==========================================
    # Membawa kedua nilai ke rentang [0, 1] agar bisa dijumlahkan secara proporsional
    norm_noise = (cost_noise - np.min(cost_noise)) / (np.max(cost_noise) - np.min(cost_noise))
    norm_lag = (cost_lag - np.min(cost_lag)) / (np.max(cost_lag) - np.min(cost_lag))

    # Cost Function Global
    J_alpha = (W_NOISE * norm_noise) + (W_LAG * norm_lag)

    # Mencari Titik Optimal (Turunan pertama = 0 -> minimum numerik)
    optimal_idx = np.argmin(J_alpha)
    optimal_alpha = alphas[optimal_idx]

    print(f"\n[HASIL KOMPUTASI OPTIMASI]")
    print(f"Titik Alpha Optimal Ideal {col_name_nice} = {optimal_alpha:.4f}")

    # ==========================================
    # 7. VISUALISASI COST FUNCTION KURVA PARABOLA
    # ==========================================
    plt.figure(figsize=(10, 6))
    plt.plot(alphas, norm_noise, label='Cost Derau (Meningkat seiring Alpha)', color='blue', linestyle='--')
    plt.plot(alphas, norm_lag, label='Cost Phase Lag (Menurun seiring Alpha)', color='red', linestyle='--')
    plt.plot(alphas, J_alpha, label='Total Cost Function $J(\\alpha)$', color='black', linewidth=2.5)

    plt.axvline(x=optimal_alpha, color='green', linestyle='-', label=f'Optimal $\\alpha$ = {optimal_alpha:.4f}')

    plt.title(f'Pemodelan Cost Function untuk Penentuan Filter EMA Optimal - {col_name_nice}')
    plt.xlabel('Nilai Parameter Filter ($\\alpha$)')
    plt.ylabel('Cost Value (Normalisasi 0-1)')
    plt.grid(True, linestyle='--', alpha=0.6)
    plt.legend()
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'cost_function_optimization.png'), dpi=300)
    print("\nPlot disimpan sebagai 'cost_function_optimization.png'")

file_path_steady = 'dataset_aht10_steady.csv'
file_path_transient = 'dataset_aht10_transient.csv'

optimize_alpha(
    file_path_steady=file_path_steady,
    file_path_transient=file_path_transient,
    target_col='temperature_c',
    output_dir='./../assets/T',
    col_name_nice='Temperature'
)

optimize_alpha(
    file_path_steady=file_path_steady,
    file_path_transient=file_path_transient,
    target_col='humidity_rh',
    output_dir='./../assets/RH',
    col_name_nice='Humidity'
)
