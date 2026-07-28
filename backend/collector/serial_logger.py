import serial

PORT = 'COM5'
BAUD = 115200
OUTPUT_FILE = 'dataset_aht10_transient.csv'

print(f"Mendengarkan data dari {PORT}...")

with serial.Serial(PORT, BAUD, timeout=1) as ser, open(OUTPUT_FILE, 'w') as f:
    try:
        while True:
            line = ser.readline().decode('utf-8', errors='ignore')
            if line:
                print(line, end='')  # Tampilkan di terminal komputer
                f.write(line)        # Tulis langsung ke file CSV secara real-time
    except KeyboardInterrupt:
        print("\nPengambilan data dihentikan. File CSV berhasil disimpan.")
        