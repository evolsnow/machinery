package machinery

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/RichardKnop/machinery/v1/backends"
	"github.com/RichardKnop/machinery/v1/brokers"
	"github.com/RichardKnop/machinery/v1/config"
)

// BrokerFactory creates a new object with brokers.Broker interface
// Currently only AMQP broker is supported
func BrokerFactory(cnf *config.Config) (brokers.Broker, error) {
	if strings.HasPrefix(cnf.Broker, "amqp://") {
		return brokers.NewAMQPBroker(cnf), nil
	}

	if strings.HasPrefix(cnf.Broker, "redis://") {

		parts := strings.Split(cnf.Broker, "redis://")
		if len(parts) != 2 {
			return nil, fmt.Errorf(
				"Redis broker connection string should be in format redis://host:port, instead got %s",
				cnf.Broker,
			)
		}

		redisHost, redisPassword, redisDB, err := parseRedisURL(cnf.Broker)
		if err != nil {
			return nil, err
		}
		return brokers.NewRedisBroker(cnf, redisHost, redisPassword, redisDB), nil
	}

	if strings.HasPrefix(cnf.Broker, "eager") {
		return brokers.NewEagerBroker(), nil
	}

	return nil, fmt.Errorf("Factory failed with broker URL: %v", cnf.Broker)
}

// BackendFactory creates a new object with backends.Backend interface
// Currently supported backends are AMQP and Memcache
func BackendFactory(cnf *config.Config) (backends.Backend, error) {
	if strings.HasPrefix(cnf.ResultBackend, "amqp://") {
		return backends.NewAMQPBackend(cnf), nil
	}

	if strings.HasPrefix(cnf.ResultBackend, "memcache://") {
		parts := strings.Split(cnf.ResultBackend, "memcache://")
		if len(parts) != 2 {
			return nil, fmt.Errorf(
				"Memcache result backend connection string should be in format memcache://server1:port,server2:port, instead got %s",
				cnf.ResultBackend,
			)
		}
		servers := strings.Split(parts[1], ",")
		return backends.NewMemcacheBackend(cnf, servers), nil
	}

	if strings.HasPrefix(cnf.ResultBackend, "redis://") {
		redisHost, redisPassword, redisDB, err := parseRedisURL(cnf.ResultBackend)
		if err != nil {
			return nil, err
		}

		return backends.NewRedisBackend(cnf, redisHost, redisPassword, redisDB), nil
	}

	if strings.HasPrefix(cnf.ResultBackend, "mongodb://") {
		return backends.NewMongodbBackend(cnf)
	}

	if strings.HasPrefix(cnf.ResultBackend, "eager") {
		return backends.NewEagerBackend(), nil
	}

	return nil, fmt.Errorf("Factory failed with result backend: %v", cnf.ResultBackend)
}

func parseRedisURL(url string) (host, password string, db int, err error) {
	// redis://pwd@host/db

	parts := strings.Split(url, "redis://")
	if parts[0] != "" {
		err = errors.New("No redis scheme found")
		return
	}
	if len(parts) != 2 {
		err = fmt.Errorf("Redis connection string should be in format redis://password@host:port/db, instead got %s", url)
		return
	}
	parts = strings.Split(parts[1], "@")
	var hostAndDB string
	if len(parts) == 2 {
		//[pwd, host/db]
		password = parts[0]
		hostAndDB = parts[1]
	} else {
		hostAndDB = parts[0]
	}
	parts = strings.Split(hostAndDB, "/")
	if len(parts) == 1 {
		//[host]
		host, db = parts[0], 0 //default redis db
	} else {
		//[host, db]
		host = parts[0]
		db, err = strconv.Atoi(parts[1])
		if err != nil {
			db, err = 0, nil //ignore err here
		}
	}
	return
}
