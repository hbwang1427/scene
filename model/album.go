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
