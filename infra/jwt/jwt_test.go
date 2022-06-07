package jwt

import (
	"testing"

	"github.com/spf13/viper"
)

func TestJWTService(t *testing.T) {
	conf := viper.New()
	conf.Set("jwt.ExpireTs", 7)
	conf.Set("jwt.secret", "xxxxx")
	jwtService, err := NewJWTService(conf)
	if err != nil {
		t.Fatal(err)
	}

	claims := Claims{
		JwtUserInfo: JwtUserInfo{UserId: "1234"},
	}
	token, err := jwtService.CreateToken(&claims)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)

	claims2, err := jwtService.ParseToken(token)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(claims2.UserId)
	t.Log(claims.UserId == claims2.UserId)
}
