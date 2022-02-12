package config

import (
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"log"
	"net"
	"time"
)

type Config struct {
	ServeRestAddress string
	DbAddress        string
	DbName           string
	DbUser           string
	DbPassword       string
	MaxConnections   int
	AcquireTimeout   int
}

func GetConnector(config *Config) (pgx.ConnPoolConfig, error) {
	databaseUri := "postgres://" + config.DbUser + ":" + config.DbPassword + "@" + config.DbAddress + "/" + config.DbName
	log.Println("Connect to db")
	//log.Println("databaseUri: " + databaseUri)
	pgxConnConfig, err := pgx.ParseURI(databaseUri)
	if err != nil {
		return pgx.ConnPoolConfig{}, errors.Wrap(err, "failed to parse database URI from environment variable")
	}
	pgxConnConfig.Dial = (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 5 * time.Minute}).Dial
	pgxConnConfig.RuntimeParams = map[string]string{
		"standard_conforming_strings": "on",
	}
	pgxConnConfig.PreferSimpleProtocol = true

	return pgx.ConnPoolConfig{
		ConnConfig:     pgxConnConfig,
		MaxConnections: config.MaxConnections,
		AcquireTimeout: time.Duration(config.AcquireTimeout) * time.Second,
	}, nil
}

func NewConnectionPool(config pgx.ConnPoolConfig) (*pgx.ConnPool, error) {
	return pgx.NewConnPool(config)
}
