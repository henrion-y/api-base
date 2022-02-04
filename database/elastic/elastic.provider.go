package elastic

import (
	"errors"
	"github.com/olivere/elastic"
	"github.com/spf13/viper"
)

func NewElasticProvider(config *viper.Viper) (*elastic.Client, error) {

	host := config.GetString("elastic.Host")
	if len(host) == 0 {
		return nil, errors.New("host is empty")
	}

	setSniff := config.GetBool("elastic.SetSniff")

	elasticClient, err := elastic.NewClient(elastic.SetSniff(setSniff), elastic.SetURL(host))
	if err != nil {
		return nil, err
	}

	return elasticClient, nil
}
