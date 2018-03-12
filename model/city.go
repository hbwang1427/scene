package model

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var citys []City

type City struct {
	CityId  uint `json:"id"`
	Name    string
	Country string
	Coord   struct {
		Lat float64
		Lng float64 `json:"lon"`
	}
}

func init() {
	if content, err := ioutil.ReadFile("assets/city.list.json"); err == nil {
		if err := json.Unmarshal(content, &citys); err != nil {
			log.Printf("unmarshal error:%v", err)
		}
		log.Printf("%d cities loaded", len(citys))
	} else {
		log.Printf("read city list error:%v", err)
	}
}

func GetCitys() []City {
	return citys
}
