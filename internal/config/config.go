package config

import (
	"log"

	"github.com/spf13/viper"
)

// Definisikan struct Config untuk menampung semua konfigurasi aplikasi
// Kita menggunakan 'mapstructure' tag agar Viper tahu ke field mana
// harus mem-parsing nilai dari file .env.
type Config struct {
	AppPort     string `mapstructure:"APP_PORT"`
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	JWTSecret   string `mapstructure:"JWT_SECRET"`
}

// LoadConfig adalah fungsi yang akan mencari dan membaca file konfigurasi.
// Ini adalah "best practice" untuk memisahkan logika loading config
// dari logika aplikasi utama.
func LoadConfig(path string) (config Config, err error) {
	// Memberitahu Viper di path mana file config berada
	viper.AddConfigPath(path)
	// Memberitahu Viper nama file config-nya (tanpa ekstensi)
	viper.SetConfigName(".env")
	// Memberitahu Viper tipe file config-nya
	viper.SetConfigType("env")

	// Memungkinkan Viper membaca environment variables juga (opsional tapi bagus)
	viper.AutomaticEnv()

	// Mencari dan membaca file config
	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("File .env tidak ditemukan, menggunakan environment variables.")
			// Jika file .env tidak ditemukan, tidak apa-apa,
			// Viper masih bisa membaca dari env var (jika di-set di server)
			err = nil // Reset error
		} else {
			// Error saat membaca file config
			log.Fatalf("Error membaca file config: %s", err)
			return
		}
	}

	// Parsing/unmarshal config yang sudah dibaca ke dalam struct Config
	// Ini adalah proses "menuangkan" data dari Viper ke struct Go kita.
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Tidak dapat unmarshal config: %s", err)
	}

	// (Opsional) Set default jika tidak ada di .env atau env var
	// Ini bagus untuk memastikan aplikasi selalu punya nilai
	if config.AppPort == "" {
		config.AppPort = "8080"
	}

	log.Println("Konfigurasi berhasil dimuat.")
	return
}