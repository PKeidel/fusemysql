package main

import (
	"database/sql"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

func getConfig() *mysql.Config {
	cfg := mysql.NewConfig()
	cfg.User, _ = os.LookupEnv("DB_USERNAME")
	cfg.Passwd, _ = os.LookupEnv("DB_PASSWORD")
	cfg.Net = "tcp"
	cfg.Addr, _ = os.LookupEnv("DB_HOST")
	cfg.DBName = "mysql"
	return cfg
}

func OpenConnection() *sql.DB {
	db, err := sql.Open("mysql", getConfig().FormatDSN())
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(1)

	return db
}

func OpenDb(dbname string) *sql.DB {
	cfg := getConfig()
	cfg.DBName = dbname
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}
