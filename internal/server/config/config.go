package config

// ServerConfig - конфигурация сервера.
type ServerConfig struct {
	// Address - адрес хоста сервера.
	Address string `env:"ADDRESS"`
	// StoreInterval - интервал сохранения метрик в файл.
	StoreInterval int64 `env:"STORE_INTERVAL"`
	// FileStoragePath - путь к файлу для сохранения метрик.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// Restore - признак необходимости восстановления данных из файла.
	Restore bool `env:"RESTORE"`
	// DatabaseDSN - строка подключения к базе данных.
	DatabaseDSN string `env:"DATABASE_DSN"`
	// Key - ключ шифрования передаваемых данных.
	Key string `env:"KEY"`
	// CryptoKey - путь до файла с приватным ключом
	CryptoKey string `env:"CRYPTO_KEY"`
}
