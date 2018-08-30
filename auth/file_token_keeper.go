package auth

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type fileTokenKeeper struct {
	sync.Mutex
	//tokens map: token -> expire time
	dumpFile           string
	tokens             map[string]time.Time
	tokenTTL           time.Duration
	tokenTTLResolution time.Duration
	deleteTokenFunc    func(string)
	donec              chan struct{}
	stopc              chan struct{}
}

func newFileTokenKeeper(deleteTokenFunc func(string), tokenTTL time.Duration, dumpFile string) *fileTokenKeeper {
	keeper := &fileTokenKeeper{
		tokens:             make(map[string]time.Time),
		dumpFile:           dumpFile,
		stopc:              make(chan struct{}),
		donec:              make(chan struct{}),
		deleteTokenFunc:    deleteTokenFunc,
		tokenTTL:           tokenTTL,
		tokenTTLResolution: 30 * time.Second,
	}

	err := keeper.load()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("load file token keeper error")
	}
	go keeper.run()

	return keeper
}

func (keeper *fileTokenKeeper) dump() {
	if v, err := json.Marshal(keeper.tokens); err == nil {
		err := ioutil.WriteFile(keeper.dumpFile, v, 0777)
		if err != nil {
			log.Debug("dump file token keeper error:%v", err)
		}
	}
}

//try load keep from disk
func (keeper *fileTokenKeeper) load() error {
	r, err := ioutil.ReadFile(keeper.dumpFile)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(r, &keeper.tokens); err != nil {
		return err
	}
	return nil
}

func (keeper *fileTokenKeeper) isValid(token string) bool {
	if expire, ok := keeper.tokens[token]; ok {
		return expire.After(time.Now())
	}
	return false
}

func (keeper *fileTokenKeeper) addToken(token string) {
	keeper.Lock()
	defer keeper.Unlock()
	keeper.tokens[token] = time.Now().Add(keeper.tokenTTL)
	keeper.dump()
}

func (keeper *fileTokenKeeper) deleteToken(token string) {
	keeper.Lock()
	defer keeper.Unlock()
	delete(keeper.tokens, token)
	keeper.dump()
}

func (keeper *fileTokenKeeper) resetTokenExpire(token string) {
	keeper.Lock()
	defer keeper.Unlock()
	if _, ok := keeper.tokens[token]; ok {
		keeper.tokens[token] = time.Now().Add(keeper.tokenTTL)
	}
	keeper.dump()
}

func (keeper *fileTokenKeeper) run() {
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
			keeper.dump()
			keeper.Unlock()
		case <-keeper.stopc:
			break
		}
	}
}

func (keeper *fileTokenKeeper) stop() {
	close(keeper.stopc)
	<-keeper.donec
}
