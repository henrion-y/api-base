package elastic

import (
	"errors"
	"github.com/olivere/elastic"
	"github.com/spf13/viper"
)

func NewElasticProvider(config *viper.Viper) (*elastic.Client, error) {

	host := config.GetString("elastic.host")
	if len(host) == 0 {
		return nil, errors.New("host is empty")
	}

	elasticClient, err := elastic.NewClient(elastic.SetURL(host))
	if err != nil {
		return nil, err
	}

	return elasticClient, nil
}
