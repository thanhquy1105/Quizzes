package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Redis  RedisConfig  `yaml:"redis"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Token  TokenConfig  `yaml:"token"`
}

type ServerConfig struct {
	TCPAddr         string        `yaml:"tcp_addr"`
	WSAddr          string        `yaml:"ws_addr"`
	GorillaWSAddr   string        `yaml:"gorilla_ws_addr"`
	HTTPAddr        string        `yaml:"http_addr"`
	RequestPoolSize int           `yaml:"request_pool_size"`
	MessagePoolSize int           `yaml:"message_pool_size"`
	RequestTimeout  time.Duration `yaml:"request_timeout"`
	MaxIdle         time.Duration `yaml:"max_idle"`
	LogDetail       bool          `yaml:"log_detail"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type MySQLConfig struct {
	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
}

type TokenConfig struct {
	SecretKey            string        `yaml:"secret_key"`
	AccessTokenDuration  time.Duration `yaml:"access_token_duration"`
	RefreshTokenDuration time.Duration `yaml:"refresh_token_duration"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := &Config{}
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
