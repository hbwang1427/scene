package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aitour/scene/model"
	"github.com/aitour/scene/web/config"

	"github.com/gin-gonic/gin"
	geo "github.com/hailocab/go-geoindex"
)

var (
	weatherApiKey     string = config.GetConfig().Options.WeatherApiKey
	currentWeatherUrl string = "http://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&APPID=%s"
	weatherForcastUrl string = "http://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&APPID=%s"
	geoCodeUrl        string = "https://%s/maps/api/geocode/json"
	googleMapApiKey   string = config.GetConfig().Options.GoogleMapApiKey //2,500 free requests per day, 50 requests per second
	googleMapDomain   string = config.GetConfig().Options.GoogleMapDomain
	cityIndex         *geo.ClusteringIndex
	museumIndex       *geo.ClusteringIndex
	citys             []model.City
	museums           []*model.Museum
)

func init() {
	citys = model.GetCitys()
	cityIndex = geo.NewClusteringIndex()
	for _, city := range citys {
		cityIndex.Add(geo.NewGeoPoint(city.Name, city.Coord.Lat, city.Coord.Lng))
	}

	museums = model.GetMuseums()
	museumIndex = geo.NewClusteringIndex()
	for _, m := range museums {
		if m.Lat != 0 || m.Lng != 0 {
			museumIndex.Add(geo.NewGeoPoint(m.Name, m.Lat, m.Lng))
		}
	}

	go scrabMuseumAddreses()
}

type weatherCacheItem struct {
	fetchTime time.Time
	content   string
}

func kelvinToCelsius(tk float32) float32 {
	return tk - 275.15
}

func parseLatLng(latlng string) (lat float64, lng float64) {
	//42.364958, -71.052768
	parts := strings.Split(latlng, ",")
	if len(parts) != 2 {
		return 0, 0
	}

	lat, _ = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	return lat, lng
}

func GeoLocationToAddress(lat float64, lng float64) (string, error) {
	//https: //maps.googleapis.com/maps/api/geocode/json?latlng=40.714224,-73.961452&key=
	req, err := http.NewRequest("GET", fmt.Sprintf(geoCodeUrl, googleMapDomain), nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("key", googleMapApiKey)
	q.Add("latlng", fmt.Sprintf("%f,%f", lat, lng))
	q.Add("language", "en")
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)

	if err != nil || resp.StatusCode != http.StatusOK {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}

func AddressToGeoLocation(address string) (string, error) {
	//https: //maps.googleapis.com/maps/api/geocode/json?latlng=40.714224,-73.961452&key=
	req, err := http.NewRequest("GET", fmt.Sprintf(geoCodeUrl, googleMapDomain), nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("key", googleMapApiKey)
	q.Add("address", address)
	q.Add("language", "en")
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)

	if err != nil || resp.StatusCode != http.StatusOK {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}

func FindNearest(indexer *geo.ClusteringIndex, lat, lng float64, km float64, n int) []geo.Point {
	return indexer.KNearest(geo.NewGeoPoint("query", lat, lng), int(n), geo.Km(km), func(p geo.Point) bool {
		return true
	})
}

func GetCurrentWeather(c *gin.Context) {
	//api.openweathermap.org/data/2.5/weather?lat=35&lon=139
	var lat, lng float64
	if latlng, ok := c.GetQuery("latlng"); ok {
		lat, lng = parseLatLng(latlng)
	} else {
		lat, _ = strconv.ParseFloat(c.Query("lat"), 2)
		lng, _ = strconv.ParseFloat(c.Query("lng"), 2)
	}

	log.Printf("query weather: lat=%f, lng=%f", lat, lng)

	resp, err := http.Get(fmt.Sprintf(currentWeatherUrl, lat, lng, weatherApiKey))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"error": "out of service",
		})
		return
	}
	defer resp.Body.Close()
	io.Copy(c.Writer, resp.Body)
}

func GetWeatherForeCast(c *gin.Context) {
	//api.openweathermap.org/data/2.5/weather?lat=35&lon=139
	var lat, lng float64
	if latlng, ok := c.GetQuery("latlng"); ok {
		lat, lng = parseLatLng(latlng)
	} else {
		lat, _ = strconv.ParseFloat(c.Query("lat"), 2)
		lng, _ = strconv.ParseFloat(c.Query("lng"), 2)
	}

	log.Printf("weather forcast: lat=%f, lng=%f", lat, lng)

	resp, err := http.Get(fmt.Sprintf(weatherForcastUrl, lat, lng, weatherApiKey))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"error": "out of service",
		})
		return
	}
	defer resp.Body.Close()
	io.Copy(c.Writer, resp.Body)
}

func GeoCodeHandler(c *gin.Context) {
	address := c.Query("address")
	if len(address) > 0 {
		body, err := AddressToGeoLocation(address)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err,
			})
			return
		}
		c.Writer.WriteString(body)
		return
	} else if _, ok := c.GetQuery("lat"); ok {
		lat, _ := strconv.ParseFloat(c.Query("lat"), 2)
		lng, _ := strconv.ParseFloat(c.Query("lng"), 2)
		body, err := GeoLocationToAddress(lat, lng)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err,
			})
			return
		}
		c.Writer.WriteString(body)
	}
}

func FindNearestCityHandler(c *gin.Context) {
	var lat, lng float64
	if latlng, ok := c.GetQuery("latlng"); ok {
		lat, lng = parseLatLng(latlng)
	} else {
		lat, _ = strconv.ParseFloat(c.Query("lat"), 2)
		lng, _ = strconv.ParseFloat(c.Query("lng"), 2)
	}
	km, _ := strconv.ParseFloat(c.Query("km"), 64)
	n, _ := strconv.ParseInt(c.Query("n"), 10, 32)
	if n == 0 {
		n = 1
	}
	log.Printf("find nearest city(lat,lng=%f,%f) km(%f)", lat, lng, km)
	start := time.Now()
	points := FindNearest(cityIndex, lat, lng, km, int(n))
	log.Printf("nearest points:%v. cost:%#v", points, time.Now().Sub(start))
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func FindNearestMuseumsHandler(c *gin.Context) {
	var lat, lng float64
	if latlng, ok := c.GetQuery("latlng"); ok {
		lat, lng = parseLatLng(latlng)
	} else {
		lat, _ = strconv.ParseFloat(c.Query("lat"), 2)
		lng, _ = strconv.ParseFloat(c.Query("lng"), 2)
	}
	km, _ := strconv.ParseFloat(c.Query("km"), 64)
	n, _ := strconv.ParseInt(c.Query("n"), 10, 32)
	if n == 0 {
		n = 100
	}
	log.Printf("find nearest museum(lat,lng=%f,%f) km(%f)", lat, lng, km)
	start := time.Now()
	points := FindNearest(museumIndex, lat, lng, km, int(n))
	log.Printf("nearest museums:%v. cost:%#v", points, time.Now().Sub(start))
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func scrabMuseumAddreses() {
	type queryResp struct {
		Results []struct {
			AddressComponents []struct {
				LongName  string `json:"long_name"`
				ShortName string `json:"short_name"`
				Types     []string
			} `json:"address_components"`

			FormattedAddress string `json:"formatted_address"`

			Geometry struct {
				Location struct {
					Lat float64
					Lng float64
				}

				LocationType string `json:"location_type"`

				ViewPort struct {
					NorthEast struct {
						Lat float64
						Lng float64
					}

					SouthWest struct {
						Lat float64
						Lng float64
					}
				}
			}

			PlaceId string `json:"place_id"`
			Types   []string
		}
		Status string
	}

	log.Printf("museum scrab begin")
	for _, m := range museums {
		if len(m.Address) == 0 || len(m.Country) == 0 {
			log.Printf("fetch %s", m.Name)
			content, err := AddressToGeoLocation(m.Name)
			if err != nil {
				log.Printf("Address to geolocation error:%v", err)
				break
			}
			var resp queryResp
			err = json.Unmarshal([]byte(content), &resp)
			if err != nil {
				log.Printf("unmarshal json error:%v", err)
				break
			}
			if resp.Status != "OK" {
				log.Printf("query status unexpect:%s", resp.Status)
				continue
			}
			m.Address = resp.Results[0].FormattedAddress
			m.PlaceId = resp.Results[0].PlaceId
			m.Lat = resp.Results[0].Geometry.Location.Lat
			m.Lng = resp.Results[0].Geometry.Location.Lng

			for _, addrComp := range resp.Results[0].AddressComponents {
				isCity, isCountry := false, false
				for _, t := range addrComp.Types {
					if t == "locality" {
						isCity = true
					} else if t == "country" {
						isCountry = true
					}
				}
				if isCity {
					m.City = addrComp.LongName
				} else if isCountry {
					m.Country = addrComp.ShortName
				}
			}
		}
		content, err := json.Marshal(museums)
		if err != nil {
			log.Printf("marshal error:%v", museums)
		}
		err = ioutil.WriteFile("assets/museums.list.json", content, 0x777)
		if err != nil {
			log.Printf("write file error:%v", err)
		}
		museumIndex.Add(geo.NewGeoPoint(m.Name, m.Lat, m.Lng))
		time.Sleep(time.Second * 3)
	}

	log.Printf("museum scrab done")
}
