package auth

import (
	"sync"
	"time"
)

type tokenKeeper struct {
	sync.Mutex
	//tokens map: token -> expire time
	tokens             map[string]time.Time
	tokenTTL           time.Duration
	tokenTTLResolution time.Duration
	deleteTokenFunc    func(string)
	donec              chan struct{}
	stopc              chan struct{}
}

func newTokenKeeper(deleteTokenFunc func(string), tokenTTL time.Duration) *tokenKeeper {
	return &tokenKeeper{
		tokens:             make(map[string]time.Time),
		stopc:              make(chan struct{}),
		donec:              make(chan struct{}),
		deleteTokenFunc:    deleteTokenFunc,
		tokenTTL:           tokenTTL,
		tokenTTLResolution: 1 * time.Second,
	}
}

func (keeper *tokenKeeper) isValid(token string) bool {
	if expire, ok := keeper.tokens[token]; ok {
		return expire.After(time.Now())
	}
	return false
}

func (keeper *tokenKeeper) addToken(token string) {
	keeper.tokens[token] = time.Now().Add(keeper.tokenTTL)
}

func (keeper *tokenKeeper) deleteToken(token string) {
	keeper.Lock()
	delete(keeper.tokens, token)
	keeper.Unlock()
}

func (keeper *tokenKeeper) resetTokenExpire(token string) {
	keeper.Lock()
	if _, ok := keeper.tokens[token]; ok {
		keeper.tokens[token] = time.Now().Add(keeper.tokenTTL)
	}
	keeper.Unlock()
}

func (keeper *tokenKeeper) run() {
	ticker := time.NewTicker(keeper.tokenTTLResolution)
	defer func() {
		ticker.Stop()
		close(keeper.donec)
	}()

	for {
		select {
		case <-ticker.C:
			keeper.Lock()
			for token, expire := range keeper.tokens {
				if time.Now().After(expire) {
					delete(keeper.tokens, token)
					keeper.deleteTokenFunc(token)
				}
			}
			keeper.Unlock()
		case <-keeper.stopc:
			break
		}
	}
}

func (keeper *tokenKeeper) stop() {
	close(keeper.stopc)
	<-keeper.donec
}
