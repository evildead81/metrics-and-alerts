package agent

// AgentConfig - конфигурация агента
type AgentConfig struct {
	// Address - адрес сервера, куда отправляются метрики.
	Address string `env:"ADDRESS" json:"address"`
	// ReportInterval - интервал сбора метрик.
	ReportInterval int64 `env:"REPORT_INTERVAL" json:"report_interval"`
	// PollInterval - интервал отправки метрик.
	PollInterval int64 `env:"POLL_INTERVAL" json:"poll_interval"`
	// Key - ключ шифрования отправляемых данных.
	Key string `env:"KEY" json:"key"`
	// RateLimit максимальное количество запросов, параллельно отправляемых на сервер.
	RateLimit int `env:"RATE_LIMIT" json:"rate_limit"`
	// CryptoKey - путь до файла с публичным ключом
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	// UseRPC определяет форму обмена данными между агентом и сервером - HTTP или gRPC
	UseRPC bool `env:"USE_RPC" json:"use_rpc"`
	// ConfigPath - путь к файлу конфигурации в формате json
	ConfigPath string `env:"CONFIG"`
}
