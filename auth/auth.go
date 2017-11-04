package auth

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrUnknownProvider  = errors.New("auth: unknown provider name")
	ErrUnknownOption    = errors.New("auth: unknown option")
	ErrUnknownJWTMethod = errors.New("auth: unknown option method")
	ErrInvalidHacKey    = errors.New("auth: invalid hmac key")
	ErrTokenExpired     = errors.New("auth: token expired")
	ErrInvalidToken     = errors.New("auth: invalid token")
)

func HashPassword(password string, salt int64) string {
	//prepare salt
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, salt)
	//permutate
	buf[0], buf[2], buf[4], buf[6], buf[7], buf[5], buf[3], buf[1] = buf[7], buf[5], buf[3], buf[1], buf[0], buf[2], buf[4], buf[6]
	//sha256(sha256(password)+salt)
	s := sha256.Sum256([]byte(password))
	buf = append(s[:], buf...)
	hashSum := sha256.Sum256(buf)
	return fmt.Sprintf("%x", hashSum)
}

func VerifyPassword(password string, salt int64, hashedPassword string) bool {
	return HashPassword(password, salt) == hashedPassword
}

type AuthInfo struct {
	UserName string
}

type TokenProvider interface {
	AssignToken(userName string) (string, error)
	RevokeToken(token string) error
	GetAuthInfo(token string) (*AuthInfo, error)
}

func CreateTokenProvider(provider string, opts map[string]interface{}) (TokenProvider, error) {
	if strings.Compare(provider, "simple") == 0 {
		return createSimpleTokenProvider(opts)
	} else if strings.Compare(provider, "jwt") == 0 {
		return createJWTTokenProvider(opts)
	}
	return nil, ErrUnknownProvider
}
