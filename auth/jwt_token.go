package auth

import (
	"fmt"
	"time"

	"github.com/aitour/scene/log"
	jwt "github.com/dgrijalva/jwt-go"
)

type jwtToken struct {
	signMethod  string
	hmacSignKey string

	keeper *tokenKeeper
}

func (p *jwtToken) GetAuthInfo(token string) (*AuthInfo, error) {
	if !p.keeper.isValid(token) {
		return nil, nil
	}

	var (
		userId int64
	)

	parsed, err := jwt.Parse(string(token), func(token *jwt.Token) (interface{}, error) {
		return []byte(p.hmacSignKey), nil
	})

	switch err.(type) {
	case nil:
		if !parsed.Valid {
			return nil, ErrInvalidToken
		}

		claims := parsed.Claims.(jwt.MapClaims)
		userId = claims["userId"].(int64)
		p.keeper.resetTokenExpire(token)
		return &AuthInfo{UserId: userId}, nil
	default:
		return nil, fmt.Errorf("failed to parse jwt token: %s", err)
	}
}

func (p *jwtToken) RevokeToken(token string) error {
	p.keeper.deleteToken(token)
	return nil
}

func (p *jwtToken) AssignToken(userId int64) (string, error) {
	tk := jwt.NewWithClaims(jwt.GetSigningMethod(p.signMethod),
		jwt.MapClaims{
			"userId": userId,
		})

	token, err := tk.SignedString([]byte(p.hmacSignKey))
	if err != nil {
		log.Debugf("failed to sign jwt token: %s", err)
		return "", err
	}

	log.Debugf("jwt token: %s", token)

	p.keeper.addToken(token)

	return token, nil
}

func createJWTTokenProvider(opts map[string]interface{}) (*jwtToken, error) {
	tokenTTL := 15 * time.Minute
	signMethod := "HS256"
	var hmacSignKey string

	for k, v := range opts {
		switch k {
		case "tokenTTL":
			tokenTTL = v.(time.Duration)
		case "method":
			signMethod = v.(string)
			if jwt.GetSigningMethod(signMethod) == nil {
				return nil, ErrUnknownJWTMethod
			}
		case "key":
			hmacSignKey = v.(string)
		}
	}

	if len(hmacSignKey) == 0 {
		return nil, ErrInvalidHacKey
	}

	tp := &jwtToken{
		signMethod:  signMethod,
		hmacSignKey: hmacSignKey,
		keeper:      newTokenKeeper(func(token string) {}, tokenTTL),
	}

	go tp.keeper.run()
	return tp, nil
}
