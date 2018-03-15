package handler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aitour/scene/model"
	"github.com/aitour/scene/web/config"

	"github.com/gin-gonic/gin"
	geo "github.com/hailocab/go-geoindex"
)

const (
	LatError float64 = 92.0
	LngError float64 = 182.0
)

var (
	weatherApiKey     string = config.GetConfig().Options.WeatherApiKey
	currentWeatherUrl string = "http://api.openweathermap.org/data/2.5/weather"
	weatherForcastUrl string = "http://api.openweathermap.org/data/2.5/forecast"
	geoPlaceUrl       string = "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
	geoPlaceDetailUrl string = "https://maps.googleapis.com/maps/api/place/details/json"
	geoPhotoUrl       string = "https://maps.googleapis.com/maps/api/place/photo"
	geoCodeUrl        string = "https://%s/maps/api/geocode/json"

	googleMapApiKey string = config.GetConfig().Options.GoogleMapApiKey //2,500 free requests per day, 50 requests per second
	googleMapDomain string = config.GetConfig().Options.GoogleMapDomain
	cityIndex       *geo.ClusteringIndex
	museumIndex     *geo.ClusteringIndex
	citys           = make(map[uint]*model.City)
	museums         []*model.Museum

	//in memory caches
	weatherCache        = &SimpleCache{}
	weatherForcastCache = &SimpleCache{}
	placeCache          = &SimpleCache{}
	placeDetailCache    = &SimpleCache{}
)

func init() {
	cityIndex = geo.NewClusteringIndex()
	cs := model.GetCitys()
	for i := 0; i < len(cs); i++ {
		citys[cs[i].CityId] = &cs[i]
		cityIndex.Add(geo.NewGeoPoint(strconv.Itoa(int(cs[i].CityId)), cs[i].Coord.Lat, cs[i].Coord.Lng))
	}
	os.MkdirAll("caches/photo", 0666)
	os.MkdirAll("caches/placedetail", 0666)
	os.MkdirAll("caches/place", 0666)
	// museums = model.GetMuseums()
	// museumIndex = geo.NewClusteringIndex()
	// for _, m := range museums {
	// 	if m.Lat != 0 || m.Lng != 0 {
	// 		museumIndex.Add(geo.NewGeoPoint(m.Name, m.Lat, m.Lng))
	// 	}
	// }

	// go scrabMuseumAddreses()
}

func kelvinToCelsius(tk float32) float32 {
	return tk - 275.15
}

func makeHttpRequest(url, method string, params map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
}

func getLatLng(c *gin.Context) (lat float64, lng float64) {
	lat, lng = LatError, LngError
	if latlng, ok := c.GetQuery("latlng"); ok && len(latlng) > 0 {
		//42.364958, -71.052768
		if parts := strings.Split(latlng, ","); len(parts) == 2 {
			lat, _ = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			lng, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		}
	} else {
		lat, _ = strconv.ParseFloat(c.Query("lat"), 2)
		lng, _ = strconv.ParseFloat(c.Query("lng"), 2)
	}
	return lat, lng
}

func GeoLocationToAddress(lat float64, lng float64, language string) (string, error) {
	//https: //maps.googleapis.com/maps/api/geocode/json?latlng=40.714224,-73.961452&key=
	req, err := http.NewRequest("GET", fmt.Sprintf(geoCodeUrl, googleMapDomain), nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("key", googleMapApiKey)
	q.Add("latlng", fmt.Sprintf("%f,%f", lat, lng))
	q.Add("language", language)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)

	if err != nil || resp.StatusCode != http.StatusOK {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}

func AddressToGeoLocation(address string, language string) (string, error) {
	//https: //maps.googleapis.com/maps/api/geocode/json?latlng=40.714224,-73.961452&key=
	req, err := http.NewRequest("GET", fmt.Sprintf(geoCodeUrl, googleMapDomain), nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("key", googleMapApiKey)
	q.Add("address", address)
	q.Add("language", language)
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

func findCity(lat, lng float64) *model.City {
	points := FindNearest(cityIndex, lat, lng, 50, 1)
	if len(points) == 0 {
		return nil
	}

	cid, _ := strconv.ParseInt(points[0].Id(), 10, 32)
	return citys[uint(cid)]
}

func GetCurrentWeather(c *gin.Context) {
	//api.openweathermap.org/data/2.5/weather?lat=35&lon=139
	lat, lng := getLatLng(c)
	language := c.DefaultQuery("language", "en")

	log.Printf("query weather: lat=%f, lng=%f, language=%s", lat, lng, language)
	city := findCity(lat, lng)
	if city != nil {
		if item, ok := weatherCache.Get(fmt.Sprintf("%d_%s", city.CityId, language)); ok {
			log.Printf("return weather from cache: %v, %s", city, language)
			io.Copy(c.Writer, strings.NewReader(item.(*SimpleCacheItem).content))
			return
		}
	}

	req, err := makeHttpRequest(currentWeatherUrl, "GET", map[string]string{
		"lat":   fmt.Sprintf("%f", lat),
		"lon":   fmt.Sprintf("%f", lng),
		"APPID": weatherApiKey,
		"lang":  language,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"error": "out of service",
		})
		return
	}
	defer resp.Body.Close()
	if city != nil {
		content, _ := ioutil.ReadAll(resp.Body)
		//expires after 10 minutes
		weatherCache.Set(fmt.Sprintf("%d_%s", city.CityId, language), &SimpleCacheItem{time.Now(), 10 * 60, string(content)})
		c.Writer.Write(content)
		return
	}

	io.Copy(c.Writer, resp.Body)
}

func GetWeatherForeCast(c *gin.Context) {
	//api.openweathermap.org/data/2.5/weather?lat=35&lon=139
	lat, lng := getLatLng(c)
	language := c.DefaultQuery("language", "en")
	log.Printf("weather forcast: lat=%f, lng=%f, language=%s", lat, lng, language)
	city := findCity(lat, lng)
	if city != nil {
		if cacheItem, ok := weatherForcastCache.Get(fmt.Sprintf("%d_%s", city.CityId, language)); ok {
			log.Printf("return weather forecast from cache:%v", city)
			io.Copy(c.Writer, strings.NewReader(cacheItem.(*SimpleCacheItem).content))
			return
		}
	}

	req, err := makeHttpRequest(weatherForcastUrl, "GET", map[string]string{
		"lat":   fmt.Sprintf("%f", lat),
		"lon":   fmt.Sprintf("%f", lng),
		"APPID": weatherApiKey,
		"lang":  language,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"error": "out of service",
		})
		return
	}
	defer resp.Body.Close()
	if city != nil {
		content, _ := ioutil.ReadAll(resp.Body)
		//expires after 10 minutes
		weatherForcastCache.Set(fmt.Sprintf("%d_%s", city.CityId, language), &SimpleCacheItem{time.Now(), 10 * 60, string(content)})
		c.Writer.Write(content)
		return
	}

	io.Copy(c.Writer, resp.Body)
}

func GeoCodeHandler(c *gin.Context) {
	address := c.Query("address")
	language := c.DefaultQuery("language", "en")
	if len(address) > 0 {
		body, err := AddressToGeoLocation(address, language)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err,
			})
			return
		}
		c.Writer.WriteString(body)
		return
	}

	lat, lng := getLatLng(c)
	if lat == LatError || lng == LngError {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid lat lng",
		})
		return
	}
	body, err := GeoLocationToAddress(lat, lng, language)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err,
		})
		return
	}
	c.Writer.WriteString(body)
}

func FindNearbyCityHandler(c *gin.Context) {
	lat, lng := getLatLng(c)
	km, _ := strconv.ParseFloat(c.Query("km"), 64)
	n, _ := strconv.ParseInt(c.Query("n"), 10, 32)
	if n == 0 {
		n = 1
	}
	log.Printf("find nearby city(lat,lng=%f,%f) km(%f)", lat, lng, km)
	points := FindNearest(cityIndex, lat, lng, km, int(n))
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func FindNearestMuseumsHandler(c *gin.Context) {
	lat, lng := getLatLng(c)
	km, _ := strconv.ParseFloat(c.Query("km"), 64)
	n, _ := strconv.ParseInt(c.Query("n"), 10, 32)
	if n == 0 {
		n = 100
	}
	log.Printf("find nearest museum(lat,lng=%f,%f) km(%f)", lat, lng, km)
	points := FindNearest(museumIndex, lat, lng, km, int(n))
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func SearchNearbyMuseumsByGoogleMap(c *gin.Context) {
	lat, lng := getLatLng(c)
	language := c.DefaultQuery("language", "en")
	city := findCity(lat, lng)
	if city != nil {
		cacheKey := fmt.Sprintf("%d_%s", city.CityId, language)
		if v, ok := placeCache.Get(cacheKey); ok {
			log.Printf("return museums from cache:%v", city)
			io.Copy(c.Writer, strings.NewReader(v.(*SimpleCacheItem).content))
			return
		}

		//try read file
		cacheFile := fmt.Sprint("caches/place/%d_%s", city.CityId, language)
		if contents, err := ioutil.ReadFile(cacheFile); err == nil {
			placeCache.Set(cacheKey, &SimpleCacheItem{time.Now(), -1, string(contents)})
			io.Copy(c.Writer, bytes.NewReader(contents))
			return
		}
	}

	radius, ok := c.GetQuery("radius")
	if !ok {
		radius = "10000"
	}
	if lat == LatError || lng == LngError {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bad request parameter",
		})
		return
	}
	req, err := http.NewRequest("GET", geoPlaceUrl, nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err,
		})
		return
	}
	q := req.URL.Query()
	q.Add("key", googleMapApiKey)
	q.Add("location", fmt.Sprintf("%f,%f", lat, lng))
	q.Add("radius", radius)
	q.Add("type", "museum")
	q.Add("language", language)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("place request error:%v", err)
		c.JSON(http.StatusOK, gin.H{
			"error": "error google map api response",
		})
		return
	}
	defer resp.Body.Close()
	if city != nil {
		content, _ := ioutil.ReadAll(resp.Body)
		placeCache.Set(city, &SimpleCacheItem{time.Now(), -1, string(content)})
		io.Copy(c.Writer, bytes.NewReader(content))
	} else {
		io.Copy(c.Writer, resp.Body)
	}
}

func GetPlacePhoto(c *gin.Context) {
	reference := c.Query("ref")
	maxwidth := c.DefaultQuery("maxwidth", "200")
	if len(reference) == 0 {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	//lookup cache
	cacheFile := fmt.Sprintf("caches/photo/%s_%s", reference, maxwidth)
	if r, err := os.Open(cacheFile); err == nil {
		log.Printf("return photo from cache:%v", cacheFile)
		io.Copy(c.Writer, r)
		r.Close()
		return
	}

	//try fetch photo from google map api
	req, err := http.NewRequest("GET", geoPhotoUrl, nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err,
		})
		return
	}
	q := req.URL.Query()
	q.Add("key", googleMapApiKey)
	q.Add("photoreference", reference)
	q.Add("maxwidth", maxwidth)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "network error",
		})
		return
	}

	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("photo fetch response:%s", content)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("error google map api response: %d", resp.StatusCode),
		})
		return
	}
	io.Copy(c.Writer, bytes.NewReader(content))
	ioutil.WriteFile(cacheFile, content, 0666)
}

func GetPlaceDetail(c *gin.Context) {
	placeId := c.Query("placeid")
	language := c.DefaultQuery("language", "en")
	if len(placeId) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parameters error",
		})
		return
	}

	//try cache
	cacheKey := fmt.Sprintf("%s_%s", placeId, language)
	if cacheItem, ok := placeDetailCache.Get(cacheKey); ok {
		log.Printf("return place detail from cache:%v", cacheKey)
		io.Copy(c.Writer, strings.NewReader(cacheItem.(*SimpleCacheItem).content))
		return
	}

	//try read disk
	cacheFile := fmt.Sprintf("caches/placedetail/%s", cacheKey)
	if content, err := ioutil.ReadFile(cacheFile); err == nil {
		log.Printf("return place detail from cache:%v", cacheFile)
		io.Copy(c.Writer, bytes.NewReader(content))
		placeDetailCache.Set(cacheKey, &SimpleCacheItem{time.Now(), -1, string(content)})
		return
	}

	//try fetch
	log.Printf("cache not found:%v", cacheKey)
	req, err := http.NewRequest("GET", geoPlaceDetailUrl, nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err,
		})
		return
	}
	q := req.URL.Query()
	q.Add("key", googleMapApiKey)
	q.Add("placeid", placeId)
	q.Add("language", language)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)

	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("error google map api response: %d", resp.StatusCode),
		})
		return
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	ioutil.WriteFile(cacheFile, content, 0666)
	io.Copy(c.Writer, bytes.NewReader(content))
	placeDetailCache.Set(cacheKey, &SimpleCacheItem{time.Now(), -1, string(content)})
}

// func scrabMuseumAddreses() {
// 	type queryResp struct {
// 		Results []struct {
// 			AddressComponents []struct {
// 				LongName  string `json:"long_name"`
// 				ShortName string `json:"short_name"`
// 				Types     []string
// 			} `json:"address_components"`

// 			FormattedAddress string `json:"formatted_address"`

// 			Geometry struct {
// 				Location struct {
// 					Lat float64
// 					Lng float64
// 				}

// 				LocationType string `json:"location_type"`

// 				ViewPort struct {
// 					NorthEast struct {
// 						Lat float64
// 						Lng float64
// 					}

// 					SouthWest struct {
// 						Lat float64
// 						Lng float64
// 					}
// 				}
// 			}

// 			PlaceId string `json:"place_id"`
// 			Types   []string
// 		}
// 		Status string
// 	}

// 	log.Printf("museum scrab begin")
// 	for _, m := range museums {
// 		if len(m.Address) == 0 || len(m.Country) == 0 {
// 			log.Printf("fetch %s", m.Name)
// 			content, err := AddressToGeoLocation(m.Name)
// 			if err != nil {
// 				log.Printf("Address to geolocation error:%v", err)
// 				break
// 			}
// 			var resp queryResp
// 			err = json.Unmarshal([]byte(content), &resp)
// 			if err != nil {
// 				log.Printf("unmarshal json error:%v", err)
// 				break
// 			}
// 			if resp.Status != "OK" {
// 				log.Printf("query status unexpect:%s", resp.Status)
// 				continue
// 			}
// 			m.Address = resp.Results[0].FormattedAddress
// 			m.PlaceId = resp.Results[0].PlaceId
// 			m.Lat = resp.Results[0].Geometry.Location.Lat
// 			m.Lng = resp.Results[0].Geometry.Location.Lng

// 			for _, addrComp := range resp.Results[0].AddressComponents {
// 				isCity, isCountry := false, false
// 				for _, t := range addrComp.Types {
// 					if t == "locality" {
// 						isCity = true
// 					} else if t == "country" {
// 						isCountry = true
// 					}
// 				}
// 				if isCity {
// 					m.City = addrComp.LongName
// 				} else if isCountry {
// 					m.Country = addrComp.ShortName
// 				}
// 			}
// 		}
// 		content, err := json.Marshal(museums)
// 		if err != nil {
// 			log.Printf("marshal error:%v", museums)
// 		}
// 		err = ioutil.WriteFile("assets/museums.list.json", content, 0x777)
// 		if err != nil {
// 			log.Printf("write file error:%v", err)
// 		}
// 		museumIndex.Add(geo.NewGeoPoint(m.Name, m.Lat, m.Lng))
// 		time.Sleep(time.Second * 3)
// 	}

// 	log.Printf("museum scrab done")
// }
