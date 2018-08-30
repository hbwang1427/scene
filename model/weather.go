package model

var Languages = map[string]string{
	"Arabic":              "ar",
	"Bulgarian":           "bg",
	"Catalan":             "ca",
	"Czech":               "cz",
	"German":              "de",
	"Greek":               "el",
	"English":             "en",
	"Persian (Farsi)":     "fa",
	"Finnish":             "fi",
	"French":              "fr",
	"Galician":            "gl",
	"Croatian":            "hr",
	"Hungarian":           "hu",
	"Italian":             "it",
	"Japanese":            "ja",
	"Korean":              "kr",
	"Latvian":             "la",
	"Lithuanian":          "lt",
	"Macedonian":          "mk",
	"Dutch":               "nl",
	"Polish":              "pl",
	"Portuguese":          "pt",
	"Romanian":            "ro",
	"Russian":             "ru",
	"Swedish":             "se",
	"Slovak":              "sk",
	"Slovenian":           "sl",
	"Spanish":             "es",
	"Turkish":             "tr",
	"Ukrainian":           "ua",
	"Vietnamese":          "vi",
	"Chinese Simplified":  "zh_cn",
	"Chinese Traditional": "zh_tw",
}

type CurrentWeather struct {
	CityId   string `json:"id"`   // City ID
	CityName string `json:"name"` //  City name
	Coord    struct {
		Lon float32 `json:"lon"` //City geo location, longitude
		Lat float32 `json:"lat"` //City geo location, latitude
	}

	Weather struct {
		Id          string //Weather condition id
		Main        string //Group of weather parameters (Rain, Snow, Extreme etc.)
		Description string //Weather condition within the group
		Icon        string //Weather icon id
	}
	Base string //Internal parameter
	Main struct {
		Temp      float32 //Temperature. Unit Default: Kelvin, Metric: Celsius, Imperial: Fahrenheit.
		Pressure  uint32  //Atmospheric pressure (on the sea level, if there is no sea_level or grnd_level data), hPa
		Humidity  uint32  //Humidity, %
		TempMin   float32 `json:"temp_min"`    //Minimum temperature at the moment. This is deviation from current temp that is possible for large cities and megalopolises geographically expanded (use these parameter optionally). Unit Default: Kelvin, Metric: Celsius, Imperial: Fahrenheit.
		TempMax   float32 `json:"temp_max"`    //Maximum temperature at the moment. This is deviation from current temp that is possible for large cities and megalopolises geographically expanded (use these parameter optionally). Unit Default: Kelvin, Metric: Celsius, Imperial: Fahrenheit.
		SeaLevel  uint32  `json:"sea_level"`   // Atmospheric pressure on the sea level, hPa
		GrndLevel uint32  `json:"grnd_level "` // Atmospheric pressure on the ground level, hPa
	}
	Wind struct {
		Speed uint32 //Wind speed. Unit Default: meter/sec, Metric: meter/sec, Imperial: miles/hour.
		Deg   uint32 //Wind direction, degrees (meteorological)
	}
	Clouds struct {
		All uint32 //Cloudiness, %
	}
	Rain struct {
		ThreeHour uint32 `json:"3h"` //Rain volume for the last 3 hours
	}
	Snow struct {
		ThreeHour uint32 `json:"3h"` //Snow volume for the last 3 hours
	}
	DateTime uint32 `json:"dt"` //Time of data calculation, unix, UTC
	Sys      struct {
		Type    int     //Internal parameter
		Id      int     //Internal parameter
		Message float32 //Internal parameter
		Country string  //Country code (GB, JP etc.)
		Sunrise uint32  // Sunrise time, unix, UTC
		Sunset  uint32  // Sunset  time, unix, UTC
	}

	Cod string `json:"cod"` //Internal parameter
}

type WeatherForcast struct {
	City struct {
		Id   string `json:"id"`   // City ID
		Name string `json:"name"` //  City name
	}

	Coord struct {
		Lon float32 `json:"lon"` //City geo location, longitude
		Lat float32 `json:"lat"` //City geo location, latitude
	}
	List []struct {
		DateTime uint32 `json:"dt"` //Time of data forecasted, unix, UTC
		Main     struct {
			Temp      float32 //Temperature. Unit Default: Kelvin, Metric: Celsius, Imperial: Fahrenheit.
			Pressure  uint32  //Atmospheric pressure (on the sea level, if there is no sea_level or grnd_level data), hPa
			Humidity  uint32  //Humidity, %
			TempMin   float32 `json:"temp_min"`    //Minimum temperature at the moment. This is deviation from current temp that is possible for large cities and megalopolises geographically expanded (use these parameter optionally). Unit Default: Kelvin, Metric: Celsius, Imperial: Fahrenheit.
			TempMax   float32 `json:"temp_max"`    //Maximum temperature at the moment. This is deviation from current temp that is possible for large cities and megalopolises geographically expanded (use these parameter optionally). Unit Default: Kelvin, Metric: Celsius, Imperial: Fahrenheit.
			SeaLevel  uint32  `json:"sea_level"`   // Atmospheric pressure on the sea level, hPa
			GrndLevel uint32  `json:"grnd_level "` // Atmospheric pressure on the ground level, hPa
		}
		Weather struct {
			Id          string //Weather condition id
			Main        string //Group of weather parameters (Rain, Snow, Extreme etc.)
			Description string //Weather condition within the group
			Icon        string //Weather icon id
		}
		Wind struct {
			Speed uint32 //Wind speed. Unit Default: meter/sec, Metric: meter/sec, Imperial: miles/hour.
			Deg   uint32 //Wind direction, degrees (meteorological)
		}
		Clouds struct {
			All uint32 //Cloudiness, %
		}
		Rain struct {
			ThreeHour uint32 `json:"3h"` //Rain volume for the last 3 hours
		}
		Snow struct {
			ThreeHour uint32 `json:"3h"` //Snow volume for the last 3 hours
		}
	}

	CalcTime uint32 `json:"dt_txt"` //Data/time of calculation, UTC
}
