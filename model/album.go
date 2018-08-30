<<<<<<< HEAD
package model

import (
	"time"
)

var (
	insertPhoto = `INSERT INTO ai_album_photo (user_id, url, memo, upload_at) values ($1, $2, $3, now()) RETURNING id`

	sqlGetUserPhotos = `SELECT * from ai_album_photo where user_id=$1`
)

type UserAlbumPhoto struct {
	Id       int64     `json:"-"`
	UserId   int64     `db:"user_id" json:"-"`
	Url      string    `json:"url"`
	Memo     string    `json:"memo"`
	UploadAt time.Time `db:"upload_at" json:"uploadat"`
}

func AddPhoto(photo *UserAlbumPhoto) error {
	row := db.QueryRow(insertPhoto, photo.UserId, photo.Url, photo.Memo)
	return row.Scan(&photo.Id)
}

func GetUserPhotos(userId int64) ([]UserAlbumPhoto, error) {
	var photos []UserAlbumPhoto
	rows, err := db.Queryx(sqlGetUserPhotos, userId)
	if err != nil {
		return photos, err
	}
	for rows.Next() {
		var photo UserAlbumPhoto
		if err = rows.StructScan(&photo); err != nil {
			return photos, err
		}
		photos = append(photos, photo)
	}
	return photos, nil
}
=======
package model

var (
	insertPhoto = `INSERT INTO ai_album_photo (user, url, memo, upload_at) values ($1, $2, $3, now()) RETURNING id`

	getUserPhotos = `SELECT * from ai_album_photo where user=$1`
)

type UserAlbumPhoto struct {
	Id     int64
	UserId int64
	Url    string
	Memo   string
}

func AddPhoto(photo *UserAlbumPhoto) error {
	row := db.QueryRow(insertPhoto, photo.UserId, photo.Url, photo.Memo)
	return row.Scan(&photo.Id)
}

func GetUserPhotos(userId int64) ([]UserAlbumPhoto, error) {
	var photos []UserAlbumPhoto
	rows, err := db.Queryx(getUserPhotos, userId)
	if err != nil {
		return photos, err
	}
	for rows.Next() {
		var photo UserAlbumPhoto
		if err = rows.StructScan(&photo); err != nil {
			return photos, err
		}
		photos = append(photos, photo)
	}
	return photos, nil
}
>>>>>>> 249b59b1b72034eb4adccebd65eb7e406909de5f
