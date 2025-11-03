package httpclient

import "time"

// Config HTTP客户端配置
type Config struct {
	BaseURL          string            `yaml:"base_url" mapstructure:"base_url"`
	Timeout          time.Duration     `yaml:"timeout" mapstructure:"timeout"`
	RetryCount       int               `yaml:"retry_count" mapstructure:"retry_count"`
	RetryWaitTime    time.Duration     `yaml:"retry_wait_time" mapstructure:"retry_wait_time"`
	RetryMaxWaitTime time.Duration     `yaml:"retry_max_wait_time" mapstructure:"retry_max_wait_time"`
	Headers          map[string]string `yaml:"headers" mapstructure:"headers"`
	Debug            bool              `yaml:"debug" mapstructure:"debug"`
	LogSlowThreshold time.Duration     `yaml:"log_slow_threshold" mapstructure:"log_slow_threshold"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Timeout:          30 * time.Second,
		RetryCount:       3,
		RetryWaitTime:    1 * time.Second,
		RetryMaxWaitTime: 5 * time.Second,
		Headers:          make(map[string]string),
		Debug:            false,
		LogSlowThreshold: 3000 * time.Millisecond, // 3秒
	}
}

// Option 客户端配置选项
type Option func(*Config)

// WithBaseURL 设置基础URL
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithRetryCount 设置重试次数
func WithRetryCount(count int) Option {
	return func(c *Config) {
		c.RetryCount = count
	}
}

// WithRetryWaitTime 设置重试等待时间
func WithRetryWaitTime(wait time.Duration) Option {
	return func(c *Config) {
		c.RetryWaitTime = wait
	}
}

// WithRetryMaxWaitTime 设置最大重试等待时间
func WithRetryMaxWaitTime(maxWait time.Duration) Option {
	return func(c *Config) {
		c.RetryMaxWaitTime = maxWait
	}
}

// WithDefaultHeaders 设置客户端默认请求头
func WithDefaultHeaders(headers map[string]string) Option {
	return func(c *Config) {
		if c.Headers == nil {
			c.Headers = make(map[string]string)
		}
		for k, v := range headers {
			c.Headers[k] = v
		}
	}
}

// WithDebug 设置调试模式
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// WithLogSlowThreshold 设置慢请求阈值
func WithLogSlowThreshold(threshold time.Duration) Option {
	return func(c *Config) {
		c.LogSlowThreshold = threshold
	}
}
