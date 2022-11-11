package kafka

import "github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util"

const (
	AcksPropertyName                             = "acks"
	AutoOffsetResetPropertyName                  = "auto.offset.reset"
	BootstrapServersPropertyName                 = "bootstrap.servers"
	EnableAutoCommitPropertyName                 = "enable.auto.commit"
	EnablePartitionEOFPropertyName               = "enable.partition.eof"
	EnableSSLCertificateVerificationPropertyName = "enable.ssl.certificate.verification"
	GoApplicationRebalanceEnablePropertyName     = "go.application.rebalance.enable"
	GroupIdPropertyName                          = "group.id"
	IsolationLevelPropertyName                   = "isolation.level"
	SASLMechanismPropertyName                    = "sasl.mechanism"
	SASLPasswordPropertyName                     = "sasl.password"
	SASLUsernamePropertyName                     = "sasl.username"
	SSLCaLocationPropertyName                    = "ssl.ca.location"
	SecurityProtocolPropertyName                 = "security.protocol"
	SessionTimeOutMsPropertyName                 = "session.timeout.ms"
	TransactionalIdPropertyName                  = "transactional.id"
	TransactionalTimeoutMsPropertyName           = "transaction.timeout.ms"
)

type ConsumerConfig struct {
	// Consumer related configs
	EnableAutoCommit   bool   `mapstructure:"enable-auto-commit" json:"enable-auto-commit" yaml:"enable-auto-commit"`
	IsolationLevel     string `mapstructure:"isolation-level" json:"isolation-level" yaml:"isolation-level"`
	MaxPollRecords     int    `mapstructure:"max-poll-records" json:"max-poll-records" yaml:"max-poll-records"`
	AutoOffsetReset    string `mapstructure:"auto-offset-reset" json:"auto-offset-reset" yaml:"auto-offset-reset"`
	SessionTimeoutMs   int    `mapstructure:"session-timeout-ms" json:"session-timeout-ms" yaml:"session-timeout-ms"`
	FetchMinBytes      int    `mapstructure:"fetch-min-bytes" json:"fetch-min-bytes" yaml:"fetch-min-bytes"`
	FetchMaxBytes      int    `mapstructure:"fetch-max-bytes" json:"fetch-max-bytes" yaml:"fetch-max-bytes"`
	Delay              int    `mapstructure:"delay" json:"delay" yaml:"delay"`
	MaxRetry           int    `mapstructure:"max-retry" json:"max-retry" yaml:"max-retry"`
	EnablePartitionEOF bool   `mapstructure:"enable-partition-eof" json:"enable-partition-eof" yaml:"enable-partition-eof"`
}

type ProducerConfig struct {
	// Producer related configs
	Acks               string `mapstructure:"acks" json:"acks" yaml:"acks"`
	MaxTimeoutMs       int    `mapstructure:"max-timeout-ms" json:"max-timeout-ms" yaml:"max-timeout-ms"`
	EnableTransactions bool   `mapstructure:"enable-transactions" json:"enable-transactions" yaml:"enable-transactions"`
}

type SSLCfg struct {
	CaLocation string `mapstructure:"ca-location" json:"ca-location" yaml:"ca-location"`
	SkipVerify bool   `json:"skip-verify" yaml:"skip-verify" mapstructure:"skip-verify"`
}

type SaslCfg struct {
	Mechanisms string `mapstructure:"mechanisms" json:"mechanisms" yaml:"mechanisms"`
	Username   string `mapstructure:"username" json:"username" yaml:"username"`
	Password   string `mapstructure:"password" json:"password" yaml:"password"`
	CaLocation string `json:"ca-location" mapstructure:"ca-location" yaml:"ca-location"`
	SkipVerify bool   `json:"skip-verify" mapstructure:"skip-verify" yaml:"skip-verify"`
}

type Config struct {
	BrokerName       string         `mapstructure:"broker-name" json:"broker-name" yaml:"broker-name"`
	BootstrapServers string         `mapstructure:"bootstrap-servers" json:"bootstrap-servers" yaml:"bootstrap-servers"`
	SecurityProtocol string         `mapstructure:"security-protocol" json:"security-protocol" yaml:"security-protocol"`
	SSL              SSLCfg         `mapstructure:"ssl" json:"ssl" yaml:"ssl"`
	SASL             SaslCfg        `mapstructure:"sasl" json:"sasl" yaml:"sasl"`
	Consumer         ConsumerConfig `mapstructure:"consumer" json:"consumer" yaml:"consumer"`
	Producer         ProducerConfig `mapstructure:"producer" json:"producer" yaml:"producer"`
	//TickInterval     string         `mapstructure:"tick-interval"`
	//Exit             struct {
	//	OnFail bool `mapstructure:"on-fail"`
	//	OnEof  bool `mapstructure:"on-eof"`
	//}
}

func (c *Config) PostProcess() error {

	c.BootstrapServers = util.ResolveConfigValue(c.BootstrapServers)
	c.SASL.Username = util.ResolveConfigValue(c.SASL.Username)
	c.SASL.Password = util.ResolveConfigValue(c.SASL.Password)
	c.SASL.Mechanisms = util.ResolveConfigValue(c.SASL.Mechanisms)

	return nil
}
