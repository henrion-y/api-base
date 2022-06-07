package qiniu

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/spf13/viper"
)

type Client struct {
	accessKey        string
	secretKey        string
	upToken          string // 缓存的token
	putPolicyExpires uint64 // token有效时间秒
	upTokenExpiresAt int    // token过期时间（秒）
	bucket           string
	cfg              storage.Config
}

func NewQiNiuClient(config *viper.Viper) (*Client, error) {
	accessKey := config.GetString("qiniu.AccessKey")
	if accessKey == "" {
		return nil, errors.New("accessKey  is empty")
	}
	secretKey := config.GetString("qiniu.SecretKey")
	if secretKey == "" {
		return nil, errors.New("secretKey  is empty")
	}
	bucket := config.GetString("qiniu.Bucket")
	if bucket == "" {
		return nil, errors.New("bucket  is empty")
	}
	putPolicyExpires := config.GetUint64("qiniu.PutPolicyExpires")
	zone := config.GetString("qiniu.Zone")
	useHTTPS := config.GetBool("qiniu.UseHTTPS")
	useCdnDomains := config.GetBool("qiniu.UseCdnDomains")
	if putPolicyExpires == 0 {
		putPolicyExpires = 7200
	}

	client := &Client{
		bucket:           bucket,
		accessKey:        accessKey,
		secretKey:        secretKey,
		upToken:          "",
		putPolicyExpires: putPolicyExpires,
		upTokenExpiresAt: 0,
	}
	client.cfg = storage.Config{UseHTTPS: useHTTPS, UseCdnDomains: useCdnDomains}
	switch zone {
	default:
		client.cfg.Zone = &storage.ZoneHuanan
	}
	return client, nil
}

func (c *Client) GetUpToken() (string, int, error) {
	putPolicy := storage.PutPolicy{
		Scope: c.bucket,
	}
	putPolicy.Expires = c.putPolicyExpires
	mac := qbox.NewMac(c.accessKey, c.secretKey)
	c.upToken = putPolicy.UploadToken(mac)
	c.upTokenExpiresAt = time.Now().Second() + int(c.putPolicyExpires)
	return c.upToken, c.upTokenExpiresAt, nil
}

func (c *Client) UploadLocalFile(localFile string, remoteFile string) (*storage.PutRet, error) {
	if time.Now().Second() > c.upTokenExpiresAt {
		_, _, err := c.GetUpToken()
		if err != nil {
			return nil, err
		}
	}

	formUploader := storage.NewFormUploader(&c.cfg)
	ret := &storage.PutRet{}
	err := formUploader.PutFile(context.Background(), &ret, c.upToken, remoteFile, localFile, nil)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) ResumeUploaderFile(localFile string, remoteFile string) (*storage.PutRet, error) {
	if time.Now().Second() > c.upTokenExpiresAt {
		_, _, err := c.GetUpToken()
		if err != nil {
			return nil, err
		}
	}

	resumeUploader := storage.NewResumeUploaderV2(&c.cfg)
	ret := &storage.PutRet{}
	recorder, err := storage.NewFileRecorder(os.TempDir())
	if err != nil {
		return nil, err
	}
	putExtra := storage.RputV2Extra{
		Recorder: recorder,
	}
	err = resumeUploader.PutFile(context.Background(), &ret, c.upToken, remoteFile, localFile, &putExtra)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
