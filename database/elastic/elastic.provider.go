package elastic

import (
	"errors"
	elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
	"github.com/spf13/viper"
)

func NewElasticProvider(config *viper.Viper) (*elasticsearch6.Client, error) {
	addresses := config.GetStringSlice("elastic.addresses")
	if len(addresses) == 0 {
		return nil, errors.New("addresses is empty")
	}

	username := config.GetString("elastic.username")
	if len(username) == 0 {
		return nil, errors.New("username is empty")
	}

	password := config.GetString("elastic.password")
	if len(password) == 0 {
		return nil, errors.New("password is empty")
	}

	elasticCfg := elasticsearch6.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}

	elasticClient, err := elasticsearch6.NewClient(elasticCfg)
	if err != nil {
		return nil, err
	}

	return elasticClient, nil
}
