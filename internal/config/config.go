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
	Storage        Storage  `yaml:"storage"`
}

const (
	DefaultPresignPutTTL = 5 * time.Minute
	DefaultPresignGetTTL = 15 * time.Minute
)

// ApplyDefaults fills in values that must always be set for the app to work.
func (c *Config) ApplyDefaults() {
	if c.Storage.PresignPutTTL <= 0 {
		c.Storage.PresignPutTTL = DefaultPresignPutTTL
	}
	if c.Storage.PresignGetTTL <= 0 {
		c.Storage.PresignGetTTL = DefaultPresignGetTTL
	}
}

type Storage struct {
	Endpoint       string        `yaml:"endpoint"`
	PublicEndpoint string        `yaml:"publicEndpoint"`
	Region         string        `yaml:"region"`
	Bucket         string        `yaml:"bucket"`
	AccessKey      string        `yaml:"accessKey"`
	SecretKey      string        `yaml:"secretKey"`
	UsePathStyle   bool          `yaml:"usePathStyle"`
	CreateBucket   bool          `yaml:"createBucket"`
	PresignPutTTL  time.Duration `yaml:"presignPutTTL"`
	PresignGetTTL  time.Duration `yaml:"presignGetTTL"`
}
