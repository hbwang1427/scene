package model

import (
	"time"

	"github.com/aitour/scene/auth"
)

var (
	createUserSql = `
	INSERT INTO "user" (name, password, salt, email, create_at) VALUES (:name, :password, :salt, :email, :create_at) RETURNING id
	`

	getUserByNameSql = `
	SELECT * FROM "user" WHERE name=$1
	`
)

type User struct {
	Id       int64
	Name     string
	Password string
	Salt     int64
	Email    string
	CreateAt time.Time `db:"create_at"`
}

func CreateUser(name string, email string, password string) (*User, error) {
	createAt := time.Now()

	//hash password
	hashPwd := auth.HashPassword(password, createAt.UnixNano())
	user := &User{
		Name:     name,
		Email:    email,
		Password: hashPwd,
		Salt:     createAt.UnixNano(),
		CreateAt: createAt,
	}
	rows, err := NamedQuery(createUserSql, user)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		rows.Scan(&user.Id)
	}
	return user, nil
}

func GetUserByName(name string) (*User, error) {
	var user User
	rows, err := QueryX(getUserByNameSql, name)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		if err := rows.StructScan(&user); err != nil {
			return nil, err
		}

	}
	return &user, nil
}
