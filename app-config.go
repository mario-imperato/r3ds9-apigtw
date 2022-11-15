package main

import (
	_ "embed"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/middleware"
	"github.com/mario-imperato/r3ds9-apigtw/linkedservices"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
)

var DefaultCfg = Config{
	Log: LogConfig{
		Level:      -1,
		EnableJSON: false,
	},
	App: AppConfig{
		Http: httpsrv.Config{
			BindAddress:     httpsrv.DefaultBindAddress,
			ListenPort:      httpsrv.DefaultListenPort,
			ShutdownTimeout: httpsrv.DefaultShutdownTimeout,
			ServerMode:      httpsrv.DefaultServerMode,
			ServerCtx: httpsrv.ServerContextCfg{
				Path: httpsrv.DefaultContextPath,
			},
		},
		MwRegistry: &middleware.MwHandlerRegistryConfig{
			ErrCfg: &middleware.ErrorHandlerConfig{
				DiscloseErrorInfo: middleware.MiddlewareErrorDefaultDiscoleInfo,
			},
			MetricsCfg: &middleware.PromHttpMetricsHandlerConfig{
				Namespace:  "r3ng",
				Subsystem:  "apigtw",
				Collectors: nil,
			},
			TraceCfg: &middleware.TracingHandlerConfig{
				Alphabet: middleware.MiddlewareTracingDefaultAlphabet,
				SpanTag:  middleware.MiddlewareTracingDefaultSpanTag,
				Header:   middleware.MiddlewareTracingDefaultHeader,
			},
		},
		Services: nil,
	},
}

type Config struct {
	Log LogConfig `yaml:"log"`
	App AppConfig `yaml:"config"`
}

type LogConfig struct {
	Level      int  `yaml:"level"`
	EnableJSON bool `yaml:"enablejson"`
}

type AppConfig struct {
	Http       httpsrv.Config                      `yaml:"http" mapstructure:"http" json:"http"`
	MwRegistry *middleware.MwHandlerRegistryConfig `yaml:"mw-handler-registry" mapstructure:"mw-handler-registry" json:"mw-handler-registry"`
	Services   *linkedservices.Config              `yaml:"linked-services" mapstructure:"linked-services" json:"linked-services"`
}

// Default config file.
//
//go:embed config.yml
var projectConfigFile []byte

const ConfigFileEnvVar = "R3NG_APIGTW_CFG_FILE_PATH"

func ReadConfig() (*Config, error) {

	configPath := os.Getenv(ConfigFileEnvVar)
	var cfgContent []byte
	var err error
	if configPath != "" {
		if _, err = os.Stat(configPath); err == nil {
			log.Info().Str("cfg-file-name", configPath).Msg("reading config")
			cfgContent, err = util.ReadFileAndResolveEnvVars(configPath)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("the %s env variable has been set but no file cannot be found at %s", ConfigFileEnvVar, configPath)
		}
	} else {
		log.Warn().Msgf("The config path variable %s has not been set. Reverting to bundled configuration", ConfigFileEnvVar)
		cfgContent = projectConfigFile

		// return nil, fmt.Errorf("the config path variable %s has not been set; please set", ConfigFileEnvVar)
	}

	cfg := DefaultCfg
	err = yaml.Unmarshal(cfgContent, &cfg)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if !cfg.Log.EnableJSON {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.SetGlobalLevel(zerolog.Level(cfg.Log.Level))

	return &cfg, nil
}
