package config

// Config содержит конфигурацию приложения.
type Config struct {
	LogsPath     string            `yaml:"logs_path" mapstructure:"logs_path"`
	PowerLogPath string            `yaml:"power_log_path,omitempty" mapstructure:"power_log_path"`
	CardsDBPath  string            `yaml:"cards_db_path" mapstructure:"cards_db_path"`
	Output       OutputConfig      `yaml:"output" mapstructure:"output"`
	Filters      FilterConfig      `yaml:"filters" mapstructure:"filters"`
	Performance  PerformanceConfig `yaml:"performance" mapstructure:"performance"`
}

// OutputConfig описывает настройки вывода.
type OutputConfig struct {
	Format      string `yaml:"format" mapstructure:"format"`
	JSONMode    string `yaml:"json_mode" mapstructure:"json_mode"`
	Destination string `yaml:"destination" mapstructure:"destination"`
	FilePath    string `yaml:"file_path,omitempty" mapstructure:"file_path"`
	Language    string `yaml:"language" mapstructure:"language"`
	Verbose     bool   `yaml:"verbose" mapstructure:"verbose"`
}

// FilterConfig описывает фильтры анализа.
type FilterConfig struct {
	GameMode []string `yaml:"game_mode" mapstructure:"game_mode"`
	MinTurns int      `yaml:"min_turns" mapstructure:"min_turns"`
	MaxTurns int      `yaml:"max_turns" mapstructure:"max_turns"`
}

// PerformanceConfig описывает базовые настройки производительности.
type PerformanceConfig struct {
	Workers    int `yaml:"workers" mapstructure:"workers"`
	BufferSize int `yaml:"buffer_size" mapstructure:"buffer_size"`
}
