package repository

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
	"unsafe"

	"github.com/ballspins/environmental-data-logger/backend/internal/model"
)

type SensorRepository struct {
	db *sql.DB
}

func NewSensorRepository(db *sql.DB) *SensorRepository {
	return &SensorRepository{db: db}
}

var bufPool = sync.Pool{
	New: func() interface{} {
		// Pre-alokasikan slice berkapasitas 64KB untuk menampung query batch besar tanpa alokasi terus-menerus
		b := make([]byte, 0, 64*1024)
		return &b
	},
}

func (r *SensorRepository) BatchInsert(ctx context.Context, dataPoints []model.SensorDataRow) error {
	if len(dataPoints) == 0 {
		return nil
	}

	// Ambil buffer dari pool dan reset panjangnya ke 0
	bufPtr := bufPool.Get().(*[]byte)
	buf := (*bufPtr)[:0]
	defer func() {
		*bufPtr = buf
		bufPool.Put(bufPtr)
	}()

	// Konstruksi query langsung secara efisien pada byte buffer untuk menghindari alokasi strings.Builder dan interface{}
	buf = append(buf, "INSERT INTO sensor_logs (time, node_id, temperature, humidity) VALUES "...)

	for i, dp := range dataPoints {
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

		if i < len(dataPoints)-1 {
			buf = append(buf, ',', ' ')
		}
	}

	// Jalankan SQL string mentah secara efisien dengan konversi 0-alokasi menggunakan unsafe.String (Go 1.20+)
	_, err := r.db.ExecContext(ctx, unsafeString(buf))
	return err
}

// unsafeString mengonversi byte slice ke string secara in-place dengan 0 alokasi heap
func unsafeString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}
