package auth

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"

	"github.com/aitour/scene/log"
)

var (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type simpleToken struct {
	sync.Mutex
	keeper *tokenKeeper

	//tokens: token -> username
	tokenLen int
	tokens   map[string]string
}

func (st *simpleToken) AssignToken(userName string) (string, error) {
	buf := make([]byte, st.tokenLen)

	for i := 0; i < st.tokenLen; i++ {
		bInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}

		buf[i] = letters[bInt.Int64()]
	}

	token := string(buf)
	st.Lock()
	st.tokens[token] = userName
	st.keeper.addToken(token)
	st.Unlock()
	return token, nil
}

func (st *simpleToken) RevokeToken(token string) error {
	st.Lock()
	st.keeper.deleteToken(token)
	delete(st.tokens, token)
	st.Unlock()
	return nil
}

func (st *simpleToken) GetAuthInfo(token string) (*AuthInfo, error) {
	if len(token) == 0 {
		return nil, nil
	}
	var authInfo *AuthInfo
	st.Lock()
	if userName, ok := st.tokens[token]; ok {
		st.keeper.resetTokenExpire(token)
		authInfo = &AuthInfo{
			UserName: userName,
		}
	}
	st.Unlock()
	return authInfo, nil
}

func (st *simpleToken) deleteToken(token string) {
	if username, ok := st.tokens[token]; ok {
		log.Debugf("deleting token %s for user %s", token, username)
		delete(st.tokens, token)
	}
}

func createSimpleTokenProvider(opts map[string]interface{}) (*simpleToken, error) {
	tokenTTL := 15 * time.Minute
	tokenLen := 16
	for key, v := range opts {
		switch key {
		case "tokenTTL":
			tokenTTL = v.(time.Duration)
		case "tokenLen":
			tokenLen = v.(int)
		default:
			return nil, ErrUnknownOption
		}
	}

	tp := &simpleToken{
		//keeper: newTokenKeeper(tp.deleteToken, tokenTTL),
		tokenLen: tokenLen,
		tokens:   make(map[string]string),
	}
	tp.keeper = newTokenKeeper(tp.deleteToken, tokenTTL)

	go tp.keeper.run()
	return tp, nil
}
