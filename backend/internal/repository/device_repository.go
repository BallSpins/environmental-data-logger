package repository

import (
	"errors"

	"github.com/ballspins/environmental-data-logger/backend/internal/model"
	"gorm.io/gorm"
)

type DeviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// RegisterOrGetID mengalokasikan ID unik (1-255) dengan jaminan ACID Transaction
func (r *DeviceRepository) RegisterOrGetID(mac string) (uint8, error) {
	var device model.RegisteredDevice

	// Memulai Transaksi ACID murni
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Lakukan Pessimistic Locking (FOR UPDATE) berdasarkan MAC Address
		// Ini mencegah race condition jika device yang sama mengirim request ganda bersamaan
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("mac_address = ?", mac).
			First(&device).Error

		// Jika MAC sudah terdaftar, langsung keluar dari transaksi (data aman)
		if err == nil {
			return nil
		}

		// Jika error-nya adalah bukan 'Record Not Found', berarti ada masalah database
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 2. Jika belum terdaftar, hitung jumlah device saat ini untuk mendapatkan nomor urut selanjutnya
		var count int64
		if err := tx.Model(&model.RegisteredDevice{}).Set("gorm:query_option", "FOR UPDATE").Count(&count).Error; err != nil {
			return err
		}

		// Validasi batas maksimum sistem biner kita (1 byte = 255)
		if count >= 255 {
			return errors.New("cluster limit reached: maximum 255 devices allowed")
		}

		// 3. Insert device baru. ID akan auto-increment dari 1 ke atas secara sekuensial
		newDevice := model.RegisteredDevice{
			MacAddress: mac,
		}
		if err := tx.Create(&newDevice).Error; err != nil {
			return err
		}

		device = newDevice
		return nil
	})

	if err != nil {
		return 0, err
	}

	return device.ID, nil
}