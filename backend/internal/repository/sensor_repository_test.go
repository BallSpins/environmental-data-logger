package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strconv"
	"testing"
	"time"

	"github.com/ballspins/environmental-data-logger/backend/internal/model"
)

type mockDriver struct{}

func (d mockDriver) Open(name string) (driver.Conn, error) {
	return mockConn{}, nil
}

type mockConn struct{}

func (c mockConn) Prepare(query string) (driver.Stmt, error) {
	return mockStmt{}, nil
}

func (c mockConn) Close() error {
	return nil
}

func (c mockConn) Begin() (driver.Tx, error) {
	return mockTx{}, nil
}

type mockStmt struct{}

func (s mockStmt) Close() error {
	return nil
}

func (s mockStmt) NumInput() int {
	return -1
}

func (s mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return mockResult{}, nil
}

func (s mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return nil, nil
}

type mockTx struct{}

func (t mockTx) Commit() error {
	return nil
}

func (t mockTx) Rollback() error {
	return nil
}

type mockResult struct{}

func (r mockResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r mockResult) RowsAffected() (int64, error) {
	return 0, nil
}

func init() {
	sql.Register("mock", mockDriver{})
}

func TestBatchInsert(t *testing.T) {
	db, err := sql.Open("mock", "any")
	if err != nil {
		t.Fatalf("Gagal membuka mock DB: %v", err)
	}
	defer db.Close()

	repo := NewSensorRepository(db)

	dataPoints := []model.SensorDataRow{
		{
			Timestamp: time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC),
			NodeID:    45,
			Temp:      25.50,
			Humi:      60.00,
		},
		{
			Timestamp: time.Date(2026, 7, 12, 12, 0, 2, 0, time.UTC),
			NodeID:    45,
			Temp:      25.60,
			Humi:      60.10,
		},
	}

	err = repo.BatchInsert(context.Background(), dataPoints)
	if err != nil {
		t.Fatalf("Gagal BatchInsert: %v", err)
	}
}

func BenchmarkBatchInsert(b *testing.B) {
	db, err := sql.Open("mock", "any")
	if err != nil {
		b.Fatalf("Gagal membuka mock DB: %v", err)
	}
	defer db.Close()

	repo := NewSensorRepository(db)

	// Membuat data sensor massal berukuran 1000 baris (setara 100 chunk biner ESP32)
	dataPoints := make([]model.SensorDataRow, 1000)
	now := time.Now()
	for i := 0; i < 1000; i++ {
		dataPoints[i] = model.SensorDataRow{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			NodeID:    uint8(i % 256),
			Temp:      25.5,
			Humi:      60.0,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = repo.BatchInsert(context.Background(), dataPoints)
	}
}

func BenchmarkBatchInsertFormatting(b *testing.B) {
	dataPoints := make([]model.SensorDataRow, 1000)
	now := time.Now()
	for i := 0; i < 1000; i++ {
		dataPoints[i] = model.SensorDataRow{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			NodeID:    uint8(i % 256),
			Temp:      25.5,
			Humi:      60.0,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Menyimulasikan logika formatting dari BatchInsert secara penuh untuk memverifikasi 0 alokasi
		bufPtr := bufPool.Get().(*[]byte)
		buf := (*bufPtr)[:0]

		buf = append(buf, "INSERT INTO sensor_logs (time, node_id, temperature, humidity) VALUES "...)

		for idx, dp := range dataPoints {
			buf = append(buf, '(')
			buf = append(buf, '\'')
			buf = dp.Timestamp.AppendFormat(buf, "2006-01-02 15:04:05")
			buf = append(buf, '\'', ',', ' ')
			buf = strconv.AppendUint(buf, uint64(dp.NodeID), 10)
			buf = append(buf, ',', ' ')
			buf = strconv.AppendFloat(buf, float64(dp.Temp), 'f', 2, 32)
			buf = append(buf, ',', ' ')
			buf = strconv.AppendFloat(buf, float64(dp.Humi), 'f', 2, 32)
			buf = append(buf, ')')

			if idx < len(dataPoints)-1 {
				buf = append(buf, ',', ' ')
			}
		}

		_ = unsafeString(buf)

		*bufPtr = buf
		bufPool.Put(bufPtr)
	}
}
