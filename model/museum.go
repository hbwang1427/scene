package model

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var museums []*Museum

type Museum struct {
	Name    string
	Lat     float64
	Lng     float64
	Address string
	City    string
	Country string
	PlaceId string
	Desc    string
}

func init() {
	if content, err := ioutil.ReadFile("assets/museums.list.json"); err == nil {
		if err := json.Unmarshal(content, &museums); err != nil {
			log.Printf("unmarshal error:%v", err)
		}
		log.Printf("%d museum loaded", len(museums))
	} else {
		log.Printf("read museum list error:%v", err)
	}
}

func GetMuseums() []*Museum {
	return museums
}
