// Package config 负责加载和解析网关的 TOML 配置文件。
package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 是网关的顶层配置，对应 configs/config.toml 的表结构。
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Log    LogConfig    `mapstructure:"log"`
	OpenAI OpenAIConfig `mapstructure:"openai"`
}

// ServerConfig 是 HTTP 服务配置。
type ServerConfig struct {
	Addr string `mapstructure:"addr"` // 监听地址，如 ":8080"
}

// LogConfig 是日志配置。
type LogConfig struct {
	Level string `mapstructure:"level"` // 日志级别：debug / info / warn / error
}

// OpenAIConfig 是 OpenAI 兼容供应商的配置，第 3 章接入 Provider 时使用。
type OpenAIConfig struct {
	BaseURL string `mapstructure:"base_url"` // 供应商 API 地址
	APIKey  string `mapstructure:"api_key"`  // 供应商 API Key
	Timeout string `mapstructure:"timeout"`  // 上游请求超时
}

// Load 读取配置文件并解析为 Config。
// 优先使用环境变量 CONFIGFILE 指定的路径，否则使用 ./configs/config.toml。
// viper 读取 TOML 时 key 不区分大小写，故结构体需用 mapstructure tag 对齐下划线键名。
func Load() (*Config, error) {
	path := os.Getenv("CONFIGFILE")
	if path != "" {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./configs")
	}
	viper.SetConfigType("toml")

	viper.SetEnvPrefix("LLMGW")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
