package auth

import (
	"errors"
	"log"
	"math/rand"
	"strings"
	"time"

	hashids "github.com/speps/go-hashids"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUnknownProvider  = errors.New("auth: unknown provider name")
	ErrUnknownOption    = errors.New("auth: unknown option")
	ErrUnknownJWTMethod = errors.New("auth: unknown option method")
	ErrInvalidHacKey    = errors.New("auth: invalid hmac key")
	ErrTokenExpired     = errors.New("auth: token expired")
	ErrInvalidToken     = errors.New("auth: invalid token")
)

var (
	defaultTokenProvider TokenProvider
)

func HashAndSalt(pwd []byte) string {

	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

func ComparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func GenRandomKey(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	keys := make([]int, n)
	for i := 0; i < n; i++ {
		keys[i] = int(r.Int31())
	}
	hd := hashids.NewData()
	hd.Salt = "(@#21233_%*&!)"
	hd.MinLength = 30
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode(keys)
	return e
}

type AuthInfo struct {
	User string
}

type TokenProvider interface {
	AssignToken(user string) (string, error)
	RevokeToken(token string) error
	GetAuthInfo(token string) (*AuthInfo, error)
}

type tokenKeeper interface {
	isValid(token string) bool
	addToken(token string)
	deleteToken(token string)
	resetTokenExpire(token string)
}

func CreateTokenProvider(provider string, opts map[string]interface{}) (TokenProvider, error) {
	if strings.Compare(provider, "simple") == 0 {
		return createSimpleTokenProvider(opts)
	} else if strings.Compare(provider, "jwt") == 0 {
		return createJWTTokenProvider(opts)
	}
	return nil, ErrUnknownProvider
}

func init() {
	defaultTokenProvider, _ = CreateTokenProvider("jwt", map[string]interface{}{
		"key":      "hmacsecretkey",
		"tokenTTL": 30 * time.Minute,
		"tokenLen": 16,
	})
}

func GetDefaultTokenProvider() TokenProvider {
	return defaultTokenProvider
}
