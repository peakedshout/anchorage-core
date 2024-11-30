package config

type LoggerConfig struct {
	Clear     bool   `json:"clear" yaml:"clear" comment:"clear old log"`
	LogLevel  string `json:"logLevel" yaml:"logLevel" comment:"log level"`
	NeedStack bool   `json:"needStack" yaml:"needStack" comment:"log stack"`
	NeedColor bool   `json:"needColor" yaml:"needColor" comment:"log color"`
	LogFile   string `json:"logFile" yaml:"logFile" comment:"log file"`
}
