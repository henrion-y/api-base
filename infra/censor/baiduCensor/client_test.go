package baiduCensor

import (
	"testing"

	"github.com/henrion-y/api-base/infra/censor"
	"github.com/spf13/viper"
)

func getCensorClient() (censor.Client, error) {
	conf := viper.New()
	conf.Set("baiduCensor.AK", "")
	conf.Set("baiduCensor.SK", "")
	return NewBaiduCensorClient(conf)
}

func TestCensorClient_CensorText(t *testing.T) {
	client, err := getCensorClient()
	if err != nil {
		t.Fatal(err)
	}

	censorResult, err := client.CensorText("花儿为什么这样红？")
	if err != nil {
		t.Log(err)
	}

	if censorResult.InterceptStatus {
		t.Log(censorResult.ReviewLabel)
	}
	t.Log(censorResult)
}

func TestCensorClient_CensorImage(t *testing.T) {
	client, err := getCensorClient()
	if err != nil {
		t.Fatal(err)
	}

	censorResult, err := client.CensorImage("")
	if err != nil {
		t.Log(err)
	}

	if censorResult.InterceptStatus {
		t.Log(censorResult.ReviewLabel)
	}
	t.Log(censorResult)
}
