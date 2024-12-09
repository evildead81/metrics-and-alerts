package agent

// AgentConfig - конфигурация агента
type AgentConfig struct {
	// Address - адрес сервера, куда отправляются метрики.
	Address string `env:"ADDRESS"`
	// ReportInterval - интервал сбора метрик.
	ReportInterval int64 `env:"REPORT_INTERVAL"`
	// PollInterval - интервал отправки метрик.
	PollInterval int64 `env:"POLL_INTERVAL"`
	// Key - ключ шифрования отправляемых данных.
	Key string `env:"KEY"`
	// RateLimit максимальное количество запросов, параллельно отправляемых на сервер.
	RateLimit int `env:"RATE_LIMIT"`
}
