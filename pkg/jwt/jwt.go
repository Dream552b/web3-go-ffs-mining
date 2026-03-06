package jwt

import (
	"errors"
	"fss-mining/pkg/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var (
	ErrTokenExpired   = errors.New("token 已过期")
	ErrTokenInvalid   = errors.New("token 无效")
	ErrTokenMalformed = errors.New("token 格式错误")
)

func Generate(userID uint, role string) (string, error) {
	cfg := config.Get().JWT
	expireAt := time.Now().Add(time.Duration(cfg.ExpireHours) * time.Hour)

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fss-mining",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func Parse(tokenStr string) (*Claims, error) {
	cfg := config.Get().JWT

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
