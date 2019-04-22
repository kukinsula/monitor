package metric

import (
	"flag"
)

var (
	DefaultDuration = 0 // Infini
	DefaultSleep    = 1

	DefaultConfig = &Config{
		Duration: DefaultDuration,
		Sleep:    DefaultSleep,
	}
)

type Config struct {
	Duration int
	Sleep    int
}

func NewConfig() *Config {
	config := DefaultConfig

	flag.IntVar(&config.Duration, "duration", DefaultDuration,
		"Monitoring duration in seconds (0 is infinite)")

	flag.IntVar(&config.Sleep, "sleep", DefaultSleep,
		"Update frequency in seconds")

	flag.Parse()

	return config
}
