package mongodb

import (
	"context"
	"crypto/tls"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"time"
)

var DefaultWriteConcern = writeconcern.New(writeconcern.WMajority())
var DefaultWriteTimeout = 60 * time.Second

type MDbLinkedService struct {
	cfg          MongoConfig
	mongoClient  *mongo.Client
	db           *mongo.Database
	writeConcern *writeconcern.WriteConcern
	writeTimeout time.Duration
}

func (lks *MDbLinkedService) Name() string {
	return lks.cfg.Name
}

func (lks *MDbLinkedService) IsConnected() bool {
	return lks.mongoClient != nil
}

func (lks *MDbLinkedService) WriteTimeout() time.Duration {
	return lks.writeTimeout
}

func NewLinkedService(cfg MongoConfig) (*MDbLinkedService, error) {
	lks := MDbLinkedService{cfg: cfg}
	return &lks, nil
}

func (mdb *MDbLinkedService) Connect(ctx context.Context) error {

	var mongoOptions = options.Client().ApplyURI(mdb.cfg.Host).
		SetMinPoolSize(uint64(mdb.cfg.Pool.MinConn)).
		SetMaxPoolSize(uint64(mdb.cfg.Pool.MaxConn)).
		SetMaxConnIdleTime(time.Duration(mdb.cfg.Pool.MaxConnectionIdleTime) * time.Millisecond).
		SetConnectTimeout(time.Duration(mdb.cfg.Pool.MaxWaitTime) * time.Millisecond)

	switch mdb.cfg.SecurityProtocol {
	case "TLS":
		log.Info().Bool("skip-verify", mdb.cfg.TLS.SkipVerify).Msg("security-protocol set to TLS....")
		tlsCfg := &tls.Config{
			InsecureSkipVerify: mdb.cfg.TLS.SkipVerify,
		}
		mongoOptions.SetTLSConfig(tlsCfg)
	case "PLAIN":
		log.Info().Str("security-protocol", mdb.cfg.SecurityProtocol).Msg("security-protocol set to PLAIN....nothing to do")
	default:
		log.Info().Str("security-protocol", mdb.cfg.SecurityProtocol).Msg("skipping mongo security-protocol settings")
	}

	/*
	 * Simple User/password authentication
	 */
	if mdb.cfg.User != "" {
		mongoOptions.SetAuth(options.Credential{
			AuthSource: mdb.cfg.DbName, Username: mdb.cfg.User, Password: mdb.cfg.Pwd,
		})
	}

	client, err := mongo.NewClient(mongoOptions)
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return err
	}

	mdb.mongoClient = client
	mdb.db = client.Database(mdb.cfg.DbName)
	mdb.writeConcern = EvalWriteConcern(mdb.cfg.WriteConcern)
	mdb.writeTimeout = DefaultWriteTimeout

	if mdb.cfg.WriteTimeout != "" {
		mdb.writeTimeout = util.ParseDuration(mdb.cfg.WriteTimeout, DefaultWriteTimeout)
	}

	return nil
}

func (m *MDbLinkedService) Disconnect(ctx context.Context) {
	if m.mongoClient != nil {
		defer m.mongoClient.Disconnect(ctx)
	}
}

func (m *MDbLinkedService) GetCollection(aCollectionId string, wcStr string) *mongo.Collection {

	w := m.writeConcern
	if wcStr != "" {
		w = EvalWriteConcern(wcStr)
	}

	for _, c := range m.cfg.Collections {
		if c.Id == aCollectionId {
			return m.db.Collection(c.Name, &options.CollectionOptions{WriteConcern: w})
		}
	}

	return nil
}
