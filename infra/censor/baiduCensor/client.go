package baiduCensor

import (
	"errors"

	"github.com/Baidu-AIP/golang-sdk/aip/censor"
	censor2 "github.com/henrion-y/api-base/infra/censor"
	"github.com/spf13/viper"
)

type CensorClient struct {
	ContentCensorClient *censor.ContentCensorClient
}

func NewBaiduCensorClient(config *viper.Viper) (censor2.Client, error) {
	ak := config.GetString("baiduCensor.AK")
	if len(ak) == 0 {
		return nil, errors.New("AK is empty")
	}
	sk := config.GetString("baiduCensor.SK")
	if len(sk) == 0 {
		return nil, errors.New("SK is empty")
	}
	contentCensorClient := censor.NewClient(ak, sk)
	censorClient := &CensorClient{
		ContentCensorClient: contentCensorClient,
	}
	return censorClient, nil
}
