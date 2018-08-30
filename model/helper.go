package model

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	dbhost = ""
	dbuser = ""
	dbpwd  = ""
	dbname = ""
)

func getDB() (*sqlx.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&client_encoding=utf-8", dbuser, dbpwd, dbhost, dbname)
	log.Printf("connstr:%s", connStr)
	return sqlx.Connect("postgres", connStr)
}

func SetDbArgs(host string, user string, pwd string, db string) {
	dbhost = host
	dbuser = user
	dbpwd = pwd
	dbname = db
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}
	return db.Exec(query, args...)
}

func NamedExec(query string, arg interface{}) (sql.Result, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}
	return db.NamedExec(query, arg)
}

func QueryX(query string, args ...interface{}) (*sqlx.Rows, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}
	return db.Queryx(query, args...)
}

func NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}
	return db.NamedQuery(query, arg)
}

func GetObject(dest interface{}, query string, args ...interface{}) error {
	db, err := getDB()
	if err != nil {
		return err
	}
	return db.Get(dest, query, args...)
}

func GetObjectList(dest interface{}, query string, args ...interface{}) error {
	db, err := getDB()
	if err != nil {
		return err
	}
	return db.Select(dest, query, args...)
}
