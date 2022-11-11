package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"strconv"
)

type CollectionCfg struct {
	Id   string
	Name string
}

type CollectionsCfg []CollectionCfg

type TLSConfig struct {
	CaLocation string `json:"ca-location" mapstructure:"ca-location" yaml:"ca-location"`
	SkipVerify bool   `json:"skip-verify" mapstructure:"skip-verify" yaml:"skip-verify"`
}

type MongoConfig struct {
	Name   string
	Host   string
	DbName string `mapstructure:"db-name" json:"db-name" yaml:"db-name"`
	User   string `mapstructure:"user" json:"user" yaml:"user"`
	Pwd    string `mapstructure:"pwd" json:"pwd" yaml:"pwd"`
	Pool   struct {
		MinConn               int `mapstructure:"min-conn" json:"min-conn" yaml:"min-conn"`
		MaxConn               int `mapstructure:"max-conn" json:"max-conn" yaml:"max-conn"`
		MaxWaitQueueSize      int `mapstructure:"max-wait-queue-size" json:"max-wait-queue-size" yaml:"max-wait-queue-size"`
		MaxWaitTime           int `mapstructure:"max-wait-time" json:"max-wait-time" yaml:"max-wait-time"`
		MaxConnectionIdleTime int `mapstructure:"max-conn-idle-time" json:"max-conn-idle-time" yaml:"max-conn-idle-time"`
		MaxConnectionLifeTime int `mapstructure:"max-conn-life-time" json:"max-conn-life-time" yaml:"max-conn-life-time"`
	}
	BulkWriteOrdered bool           `mapstructure:"bulk-write-ordered" json:"bulk-write-ordered" yaml:"bulk-write-ordered"`
	WriteConcern     string         `mapstructure:"write-concern" json:"write-concern" yaml:"write-concern"`
	WriteTimeout     string         `mapstructure:"write-timeout" json:"write-timeout" yaml:"write-timeout"`
	Collections      CollectionsCfg `mapstructure:"collections" json:"collections" yaml:"collections"`
	SecurityProtocol string         `json:"security-protocol" json:"security-protocol" yaml:"security-protocol"`
	TLS              TLSConfig      `json:"tls" mapstructure:"tls" yaml:"tls"`
}

/*
 * GetConfigDefaults not applicable in an array type of config.

func GetConfigDefaults() []configuration.VarDefinition {
	return []configuration.VarDefinition{
		{"config.mongo.host", "mongodb://localhost:27017", "host reference"},
		{"config.mongo.db-name", "cpxstore", "database name"},
		{"config.mongo.bulk-write-ordered", true, "bulk write ordered"},
		{"config.mongo.write-concern", "majority", "write concern"},
		{"config.mongo.write-timeout", "120s", "write timeout"},
		{"config.mongo.pool.min-conn", "1", "min conn"},
		{"config.mongo.pool.max-conn", "20", "max conn"},
		{"config.mongo.pool.max-wait-queue-size", "1000", "max wait queue size"},
		{"config.mongo.pool.max-wait-time", "1000", "max wait time"},
		{"config.mongo.pool.max-conn-idle-time", "30000", "max conn idle time"},
		{"config.mongo.pool.max-conn-life-time", "6000000", "max conn life time"},
	}
}
*/

func EvalWriteConcern(wstr string) *writeconcern.WriteConcern {

	w := DefaultWriteConcern
	if wstr != "" {
		switch wstr {
		case "majority":
			writeconcern.New(writeconcern.WMajority())
		case "1":
			w = writeconcern.New(writeconcern.W(1))
		default:
			if i, err := strconv.Atoi(wstr); err == nil {
				w = writeconcern.New(writeconcern.W(i))
			}
		}
	}

	return w

}

func (c *MongoConfig) PostProcess() error {
	return nil
}
