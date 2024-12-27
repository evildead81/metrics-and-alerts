package config

// ServerConfig - конфигурация сервера.
type ServerConfig struct {
	// Address - адрес хоста сервера.
	Address string `env:"ADDRESS" json:"address"`
	// StoreInterval - интервал сохранения метрик в файл.
	StoreInterval int64 `env:"STORE_INTERVAL" json:"store_interval"`
	// FileStoragePath - путь к файлу для сохранения метрик.
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	// Restore - признак необходимости восстановления данных из файла.
	Restore bool `env:"RESTORE" json:"restore"`
	// DatabaseDSN - строка подключения к базе данных.
	DatabaseDSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// Key - ключ шифрования передаваемых данных.
	Key string `env:"KEY" json:"key"`
	// CryptoKey - путь до файла с приватным ключом
	CryptoKey  string `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigPath string `env:"CONFIG"`
}
