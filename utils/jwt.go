package utils

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"time"
)

type AuthService interface {
	CreateToken(claims *Claims) (string, error)
	ParseToken(signedToken string) (*Claims, error)
}

type JwtUserInfo struct {
	UserId   string `json:"user_id"`   // 用户id
	UserType string `json:"user_type"` // 用户类型
}

// Claims custom token
type Claims struct {
	JwtUserInfo
	jwt.StandardClaims
}

type JWTService struct {
	Secret   string        `json:"Secret"`
	ExpireTs time.Duration `json:"ExpireTs"`
}

func NewJWTService(config *viper.Viper) (AuthService, error) {
	jWTService := &JWTService{}
	jWTService.Secret = config.GetString("jwt.Secret")
	if len(jWTService.Secret) == 0 {
		return nil, errors.New("secret  is empty")
	}
	expireTs := config.GetInt64("jwt.ExpireTs")
	if expireTs == 0 {
		expireTs = 72
	}
	jWTService.ExpireTs = time.Duration(expireTs)
	return jWTService, nil
}

// CreateToken create token
func (s *JWTService) CreateToken(claims *Claims) (string, error) {
	claims.ExpiresAt = time.Now().Add(time.Hour * s.ExpireTs).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.Secret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// ParseToken validate token
func (s *JWTService) ParseToken(signedToken string) (*Claims, error) {
	// todo 这里要考虑拿到过期时间做刷新
	token, err := jwt.ParseWithClaims(signedToken, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected login method %v", token.Header["alg"])
			}
			return []byte(s.Secret), nil
		})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("校验失败")
}
