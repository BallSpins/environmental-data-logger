import pandas as pd
import numpy as np

# EMA Function
def calculate_ema(data, alpha):
    ema = np.zeros(len(data))
    ema[0] = data.iloc[0]
    for i in range(1, len(data)):
        ema[i] = alpha * data.iloc[i] + (1 - alpha) * ema[i-1]
    return ema

def analyse_rmse(file_path, target_col, unit_label, alphas):
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
	for alpha in alphas:
	    df[f'ema_{alpha}'] = calculate_ema(df[target_col], alpha)

	# Filter for transient window: minute 3.0 to 6.0
	window_df = df[(df['time_min'] >= 3.0) & (df['time_min'] <= 6.0)]

	# Calculate Metrics
	results = []
	for alpha in alphas:
		raw = window_df[target_col]
		ema = window_df[f'ema_{alpha}']
	
	    # RMSE
		rmse = np.sqrt(np.mean((raw - ema)**2))
	
	    # Max Lag Error (Absolute)
		max_error = np.max(np.abs(raw - ema))
	
	    # Time of Max Error
		idx_max = np.argmax(np.abs(raw - ema))
		time_max_error = window_df['time_min'].iloc[idx_max]
	
		results.append({
		    'Alpha (α)': alpha,
		    f'RMSE ({unit_label})': round(rmse, 4),
		    f'Max Lag Error ({unit_label})': round(max_error, 4),
		    'Waktu Max Error (Menit)': round(time_max_error, 2)
		})

	results_df = pd.DataFrame(results)
	print(results_df.to_markdown(index=False), '\n')

file_path = 'dataset_aht10_transient.csv'

analyse_rmse(
	file_path=file_path,
	target_col='temperature_c',
	unit_label='°C',
	alphas = [0.05, 0.0666, 0.1, 0.1494, 0.2, 0.4]
)

analyse_rmse(
	file_path=file_path,
	target_col='humidity_rh',
	unit_label='%RH',
	alphas = [0.05, 0.0666, 0.1, 0.2260, 0.2, 0.4]
)
