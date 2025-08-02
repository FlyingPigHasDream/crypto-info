package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	App        App        `mapstructure:"app"`
	Server     Server     `mapstructure:"server"`
	Log        Log        `mapstructure:"log"`
	Database   Database   `mapstructure:"database"`
	ExternalAPI ExternalAPI `mapstructure:"external_api"`
	Cache      Cache      `mapstructure:"cache"`
	Monitoring Monitoring `mapstructure:"monitoring"`
	RateLimit  RateLimit  `mapstructure:"rate_limit"`
	Security   Security   `mapstructure:"security"`
	Business   Business   `mapstructure:"business"`
}

// App 应用配置
type App struct {
	Name     string `mapstructure:"name"`
	Version  string `mapstructure:"version"`
	Env      string `mapstructure:"env"`
	Debug    bool   `mapstructure:"debug"`
	Timezone string `mapstructure:"timezone"`
}

// Server 服务器配置
type Server struct {
	HTTP HTTPServer `mapstructure:"http"`
	GRPC GRPCServer `mapstructure:"grpc"`
}

// HTTPServer HTTP服务器配置
type HTTPServer struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// GRPCServer GRPC服务器配置
type GRPCServer struct {
	Host    string        `mapstructure:"host"`
	Port    int           `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// Log 日志配置
type Log struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// Database 数据库配置
type Database struct {
	Redis RedisConfig `mapstructure:"redis"`
	MySQL MySQLConfig `mapstructure:"mysql"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parse_time"`
	Loc             string        `mapstructure:"loc"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// ExternalAPI 外部API配置
type ExternalAPI struct {
	Huobi   APIConfig `mapstructure:"huobi"`
	Binance APIConfig `mapstructure:"binance"`
}

// APIConfig API配置
type APIConfig struct {
	BaseURL       string        `mapstructure:"base_url"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryTimes    int           `mapstructure:"retry_times"`
	RetryInterval time.Duration `mapstructure:"retry_interval"`
}

// Cache 缓存配置
type Cache struct {
	PriceTTL   time.Duration `mapstructure:"price_ttl"`
	VolumeTTL  time.Duration `mapstructure:"volume_ttl"`
	DefaultTTL time.Duration `mapstructure:"default_ttl"`
}

// Monitoring 监控配置
type Monitoring struct {
	Metrics     MetricsConfig     `mapstructure:"metrics"`
	Tracing     TracingConfig     `mapstructure:"tracing"`
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Port    int    `mapstructure:"port"`
}

// TracingConfig 链路追踪配置
type TracingConfig struct {
	Enabled        bool    `mapstructure:"enabled"`
	JaegerEndpoint string  `mapstructure:"jaeger_endpoint"`
	ServiceName    string  `mapstructure:"service_name"`
	SampleRate     float64 `mapstructure:"sample_rate"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Path     string        `mapstructure:"path"`
	Interval time.Duration `mapstructure:"interval"`
}

// RateLimit 限流配置
type RateLimit struct {
	Enabled           bool          `mapstructure:"enabled"`
	RequestsPerSecond int           `mapstructure:"requests_per_second"`
	Burst             int           `mapstructure:"burst"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

// Security 安全配置
type Security struct {
	CORS CORSConfig `mapstructure:"cors"`
	JWT  JWTConfig  `mapstructure:"jwt"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	ExpireTime time.Duration `mapstructure:"expire_time"`
	Issuer     string        `mapstructure:"issuer"`
}

// Business 业务配置
type Business struct {
	SupportedSymbols    []string `mapstructure:"supported_symbols"`
	DefaultSymbol       string   `mapstructure:"default_symbol"`
	MaxAnalysisDays     int      `mapstructure:"max_analysis_days"`
	DefaultAnalysisDays int      `mapstructure:"default_analysis_days"`
	MockDataEnabled     bool     `mapstructure:"mock_data_enabled"`
}

// Load 加载配置
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 默认配置文件路径
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
		v.AddConfigPath("/etc/crypto-info")
	}

	// 环境变量支持
	v.SetEnvPrefix("CRYPTO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 根据环境加载对应的配置文件
	env := v.GetString("app.env")
	if env != "" {
		envConfigPath := filepath.Join(filepath.Dir(v.ConfigFileUsed()), env+".yaml")
		if _, err := os.Stat(envConfigPath); err == nil {
			v.SetConfigFile(envConfigPath)
			if err := v.MergeInConfig(); err != nil {
				return nil, fmt.Errorf("failed to merge env config: %w", err)
			}
		}
	}

	// 解析配置
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validate 验证配置
func validate(config *Config) error {
	if config.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	if config.Server.HTTP.Port <= 0 || config.Server.HTTP.Port > 65535 {
		return fmt.Errorf("invalid http port: %d", config.Server.HTTP.Port)
	}

	if config.Server.GRPC.Port <= 0 || config.Server.GRPC.Port > 65535 {
		return fmt.Errorf("invalid grpc port: %d", config.Server.GRPC.Port)
	}

	return nil
}

// GetHTTPAddr 获取HTTP服务地址
func (c *Config) GetHTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.HTTP.Host, c.Server.HTTP.Port)
}

// GetGRPCAddr 获取GRPC服务地址
func (c *Config) GetGRPCAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.GRPC.Host, c.Server.GRPC.Port)
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}