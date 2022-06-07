package qiniu

import (
	"testing"

	"github.com/spf13/viper"
)

func getClient(t *testing.T) *Client {
	conf := viper.New()
	client, err := NewQiNiuClient(conf)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestClient_GetUpToken(t *testing.T) {
	client := getClient(t)
	token, i, err := client.GetUpToken()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)
	t.Log(i)
}

func TestClient_UploadLocalFile(t *testing.T) {
	client := getClient(t)
	rest, err := client.UploadLocalFile(``, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rest)
}

func TestClient_ResumeUploaderFile(t *testing.T) {
	client := getClient(t)
	rest, err := client.ResumeUploaderFile(``, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rest)
}
