package restclient

import (
	"github.com/opentracing/opentracing-go"
	"time"
)

type Config struct {
	RestTimeout time.Duration `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
	SkipVerify  bool          `mapstructure:"skip-verify" json:"skip-verify" yaml:"skip-verify"`

	Span           opentracing.Span
	TraceOpName    string `mapstructure:"trace-op-name" json:"trace-op-name" yaml:"trace-op-name"`
	NestTraceSpans bool   `mapstructure:"nest-trace-ops" json:"nest-trace-ops" yaml:"nest-trace-ops"`
}

type Option func(o *Config)

func WithTraceOperationName(opn string) Option {
	return func(o *Config) {
		o.TraceOpName = opn
	}
}

func WithSkipVerify(b bool) Option {
	return func(o *Config) {
		o.SkipVerify = b
	}
}

func WithTimeout(to time.Duration) Option {
	return func(o *Config) {
		o.RestTimeout = to
	}
}
