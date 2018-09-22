package model

const (
	insert_porcelain = `INSERT INTO ai_porcelain_photo (id, url, class, upload_at, upload_from, file_hash) VALUES ($1, $2, $3, now(), $4, $5)`

	update_porcelain_class = `UPDATE ai_porcelain_photo SET class=$2 WHERE id=$1`
)

func InsertPorcelainPhoto(id int64, url string, class int, uploadFrom string, fileHash string) error {
	_, err := db.Exec(insert_porcelain, id, url, class, uploadFrom, fileHash)
	return err
}

func UpdatePorcelainClass(id int64, class uint16) error {
	_, err := db.Exec(update_porcelain_class, id, class)
	return err
}
