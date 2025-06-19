package config

import "time"

type Config struct {
	Port int    `yaml:"port"`
	Env  string `yaml:"env"`
	Db   struct {
		Dsn          string        `yaml:"dsn"`
		MaxOpenConns int           `yaml:"maxOpenConns"`
		MaxIdleConns int           `yaml:"maxIdleConns"`
		MaxIdleTime  time.Duration `yaml:"maxIdleTime"`
	} `yaml:"db"`
	TrustedOrigins []string `yaml:"trustedOrigins"`
}
