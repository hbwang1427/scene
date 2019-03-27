package model

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	sqlGetArtReferences = "select image_id, art_id, image_location, image_feature, mobilenet_feature from ai_reference_image"
	sqlGetArt           = `select a.art_id, a.museum_id, a.artist_id, a.display_number, a.creation_year, a.price,
	b.image_location, c.title, c.location, c.material, c.category, c.text, c.audio, 
	d.name as museum_name, d.city as museum_city, d.country as museum_country
	 from ai_art a, ai_reference_image b, ai_art_information c, ai_museum_information d
	 where a.art_id=b.art_id and a.art_id=c.art_id and a.museum_id = d.museum_id and a.art_id=$1 and c.language_id=$2`
	artReferences []ArtReference = nil
)

type ArtReference struct {
	ImageID       int
	ArtID         int
	ImageLocation string
	ImageFeature  []float64 //already normalized feature
	//ImageFeatureNorm     float64
	MobileNetFeature []float64 //already normalized feature
	//MobileNetFeatureNorm float64
}

func Norm(array []float64) float64 {
	if len(array) == 0 {
		return 0
	}
	var avg float64
	for _, v := range array {
		avg += v
	}
	avg = avg / float64(len(array))
	var n float64
	for _, v := range array {
		n += (v - avg) * (v - avg)
	}
	return float64(math.Sqrt(float64(n)))
}

func GetArtReferences() ([]ArtReference, error) {
	if artReferences != nil {
		return artReferences, nil
	}

	return nil, fmt.Errorf("references loading....")
}

func loadArtReferences() ([]ArtReference, error) {
	var references []ArtReference
	var gobFile = "./artref.gob"
	if file, err := os.Open(gobFile); err == nil {
		decoder := gob.NewDecoder(file)
		decoder.Decode(&references)
		file.Close()
		log.Printf("load image references count: %v", len(references))
		return references, nil
	}

	rows, err := db.Queryx(sqlGetArtReferences)
	if err != nil {
		return artReferences, err
	}
	for rows.Next() {
		var ref ArtReference
		var imageFeature, mobileNetFeature string
		if err := rows.Scan(&ref.ImageID, &ref.ArtID, &ref.ImageLocation, &imageFeature, &mobileNetFeature); err != nil {
			return nil, err
		}

		if imageFeature[0] == '{' {
			imageFeature = imageFeature[1 : len(imageFeature)-1]
		}
		if mobileNetFeature[0] == '{' {
			mobileNetFeature = mobileNetFeature[1 : len(mobileNetFeature)-1]
		}
		imageFeature = strings.Replace(imageFeature, "\n", "", -1)
		imageFeature = strings.Replace(imageFeature, "\"", "", -1)
		mobileNetFeature = strings.Replace(mobileNetFeature, "\n", "", -1)
		mobileNetFeature = strings.Replace(mobileNetFeature, "\"", "", -1)
		for _, f := range strings.Split(imageFeature, ",") {
			if v, err := strconv.ParseFloat(f, 32); err != nil {
				return nil, err
			} else {
				ref.ImageFeature = append(ref.ImageFeature, float64(v))
			}
		}
		for _, f := range strings.Split(mobileNetFeature, ",") {
			if v, err := strconv.ParseFloat(f, 32); err != nil {
				return nil, err
			} else {
				ref.MobileNetFeature = append(ref.MobileNetFeature, float64(v))
			}

		}
		//ref.ImageFeatureNorm = Norm(ref.ImageFeature)
		//ref.MobileNetFeatureNorm = Norm(ref.MobileNetFeature)
		references = append(references, ref)
	}

	if file, err := os.Create(gobFile); err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(references)
		file.Close()
	}
	return references, nil
}

type Art struct {
	ArtID         int
	MuseumID      int
	ArtistID      int
	DisplayNumber int
	CreationYear  string
	Price         int
	Title         string
	Category      string
	Location      string
	Images        []string
	Audios        []string
	Text          string
	Material      string
	MuseumName    string
	MuseumCity    string
	MuseumCountry string
}

func GetArtById(artId int, language_id int) (*Art, error) {
	row := db.QueryRowx(sqlGetArt, artId, language_id)
	var art Art
	var imageUrl, audioUrl sql.NullString
	var displayNumber, price sql.NullInt64
	if err := row.Scan(&art.ArtID, &art.MuseumID, &art.ArtistID, &displayNumber, &art.CreationYear, &price,
		&imageUrl, &art.Title, &art.Location, &art.Material, &art.Category, &art.Text, &audioUrl,
		&art.MuseumName, &art.MuseumCity, &art.MuseumCountry); err != nil {
		return nil, err
	}

	art.Images = append(art.Images, imageUrl.String)
	art.Audios = append(art.Audios, audioUrl.String)
	art.DisplayNumber = int(displayNumber.Int64)
	art.Price = int(price.Int64)
	return &art, nil
}
