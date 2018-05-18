package model

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/aitour/scene/web/config"

	"github.com/aitour/scene/auth"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	sqlCreateUser = `
	INSERT INTO "ai_user" (email, password, create_at, activate_key, enabled, deleted) VALUES ($1, $2, now(), $3, true, false) RETURNING id
	`

	sqlCheckActivationKey = `
	SELECT COUNT(*) FROM ai_user WHERE email=$1 AND activate_key=$2 AND EXTRACT(EPOCH FROM (timestamp 'now' - create_at))/3600 < 24
	`

	sqlActivateUserAccount = `
	UPDATE ai_user set activated=true, activate_key='' where email=$1
	`

	sqlGetActivatedUser = `
	SELECT name, phone, create_at, password FROM "ai_user" WHERE email=$1 and enabled=true and activated=true and deleted=false
	`
)

var (
	ErrorInvalidUserNameOrPassword = fmt.Errorf("invalid user name or password")
)

var (
	cfg *config.Config
	db  *sqlx.DB
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
	} else {
		log.Printf("db connected")
	}
}

type User struct {
	Id          int64
	Name        string
	Phone       string
	Password    string
	Email       string
	ActivateKey string    `db:"activate_key"`
	CreateAt    time.Time `db:"create_at"`
}

func CreateUser(email string, password string) (*User, error) {
	createAt := time.Now()

	//hash password
	hashPwd := auth.HashAndSalt([]byte(password))
	user := &User{
		Email:       email,
		Password:    hashPwd,
		CreateAt:    createAt,
		ActivateKey: auth.GenRandomKey(8),
	}

	row := db.QueryRow(sqlCreateUser, user.Email, user.Password, user.ActivateKey)
	if row == nil {
		return nil, fmt.Errorf("query error")
	}
	err := row.Scan(&user.Id)
	return user, err
}

func VerifyUser(email string, password string) (*User, error) {
	var user User
	row := db.QueryRowx(sqlGetActivatedUser, email)
	var name, phone sql.NullString
	if err := row.Scan(&name, &phone, &user.CreateAt, &user.Password); err != nil {
		return nil, err
	}

	checkPass := auth.ComparePasswords(user.Password, []byte(password))
	if !checkPass {
		return nil, ErrorInvalidUserNameOrPassword
	}
	user.Password = ""
	user.Name = name.String
	user.Phone = name.String
	return &user, nil
}

func ActivateUser(email string, activationKey string) bool {
	var count int64
	if err := db.Get(&count, sqlCheckActivationKey, email, activationKey); err != nil {
		return false
	}

	result, err := db.Exec(sqlActivateUserAccount, email)
	if err != nil {
		return false
	}
	count, err = result.RowsAffected()
	return err == nil && count == 1
}
