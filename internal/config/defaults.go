package config

const (
	DefaultLogsPath          = "logs"
	DefaultCardsDBPath       = "data/cards.json"
	DefaultOutputFormat      = "text"
	DefaultJSONMode          = "replay"
	DefaultOutputDestination = "stdout"
	DefaultOutputLanguage    = "ru"
	DefaultPerformanceWorker = 1
	DefaultBufferSize        = 64 * 1024
)

// DefaultConfig возвращает конфигурацию приложения по умолчанию.
func DefaultConfig() Config {
	return Config{
		LogsPath:    DefaultLogsPath,
		CardsDBPath: DefaultCardsDBPath,
		Output: OutputConfig{
			Format:      DefaultOutputFormat,
			JSONMode:    DefaultJSONMode,
			Destination: DefaultOutputDestination,
			Language:    DefaultOutputLanguage,
			Verbose:     false,
		},
		Filters: FilterConfig{
			GameMode: []string{"standard"},
			MinTurns: 0,
			MaxTurns: 0,
		},
		Performance: PerformanceConfig{
			Workers:    DefaultPerformanceWorker,
			BufferSize: DefaultBufferSize,
		},
	}
}
