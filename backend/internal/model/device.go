package model

import "time"

type RegisteredDevice struct {
	ID					uint8			`gorm:"primaryKey;autoIncrement"`
	MacAddress	string		`gorm:"type:varchar(17);uniqueIndex;not null"`
	CreatedAt		time.Time
}

type SensorDataRow struct {
	Timestamp time.Time `gorm:"column:time;type:datetime;not null;index"`
	NodeID    uint8     `gorm:"column:node_id;type:tinyint unsigned;not null;index"`
	Temp      float32   `gorm:"column:temperature;type:float;not null"`
	Humi      float32   `gorm:"column:humidity;type:float;not null"`
}

func (SensorDataRow) TableName() string {
	return "sensor_logs"
}

// 4 byte
type SensorData struct {
	Humi	int16	// 2 byte
	Temp	int16	// 2 byte
}

// 48 byte
type ChunkPayload struct {
	Timestamp uint32					// Byte-0: timestamp (4 byte)
	NodeID		uint8						// Byte-1: ID (1 byte)
	Padding		[3]uint8				// Byte-2: Explicit padding for alignment memory 16-bit (3 byte)
	Data			[10]SensorData	// Byte-3 until Byte-47: 10 data from sensor (4 * 10 byte)
}