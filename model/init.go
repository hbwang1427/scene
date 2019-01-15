package model

import (
	"fmt"
	"sync"

	"github.com/aitour/scene/web/config"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

var (
	cfg *config.Config
	db  *sqlx.DB

	artReferencesLoadOnce sync.Once
)

func init() {
	cfg = config.GetConfig()
	if cfg == nil {
		log.Fatalln("unable to get app config")
	}
	var err error
	db, err = sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Db.Host, cfg.Db.Port, cfg.Db.User, cfg.Db.Password, cfg.Db.DbName))
	if err != nil {
		log.Fatalln(err)
		//log.Println(err)
	} else {
		log.Printf("db connected")
	}


	//preload art references
	artReferencesLoadOnce.Do(func() {
		var err error
		artReferences, err = loadArtReferences()
		if err != nil {
			log.Printf("load art references error:%v", err)
		}
	})
}
