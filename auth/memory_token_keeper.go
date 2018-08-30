package auth

import (
	"sync"
	"time"
)

type memoryTokenKeeper struct {
	sync.Mutex
	//tokens map: token -> expire time
	tokens             map[string]time.Time
	tokenTTL           time.Duration
	tokenTTLResolution time.Duration
	deleteTokenFunc    func(string)
	donec              chan struct{}
	stopc              chan struct{}
}

func newMemoryTokenKeeper(deleteTokenFunc func(string), tokenTTL time.Duration) *memoryTokenKeeper {
	keeper := &memoryTokenKeeper{
		tokens:             make(map[string]time.Time),
		stopc:              make(chan struct{}),
		donec:              make(chan struct{}),
		deleteTokenFunc:    deleteTokenFunc,
		tokenTTL:           tokenTTL,
		tokenTTLResolution: 1 * time.Second,
	}

	go keeper.run()
	return keeper
}

func (keeper *memoryTokenKeeper) isValid(token string) bool {
	if expire, ok := keeper.tokens[token]; ok {
		return expire.After(time.Now())
	}
	return false
}

func (keeper *memoryTokenKeeper) addToken(token string) {
	keeper.tokens[token] = time.Now().Add(keeper.tokenTTL)
}

func (keeper *memoryTokenKeeper) deleteToken(token string) {
	keeper.Lock()
	delete(keeper.tokens, token)
	keeper.Unlock()
}

func (keeper *memoryTokenKeeper) resetTokenExpire(token string) {
	keeper.Lock()
	if _, ok := keeper.tokens[token]; ok {
		keeper.tokens[token] = time.Now().Add(keeper.tokenTTL)
	}
	keeper.Unlock()
}

func (keeper *memoryTokenKeeper) run() {
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

func (keeper *memoryTokenKeeper) stop() {
	close(keeper.stopc)
	<-keeper.donec
}
