package res

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func GetDb(host, name, user, password string) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&maxAllowedPacket=1073741824&multiStatements=true&parseTime=true&loc=Local", user, password, host, name)
	_db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open db:%s error:%+v", name, err)
	}
	if _err := _db.Ping(); _err != nil {
		log.Fatalf("db:%s is not ready now, error:%+v", name, _err)
	}
	_db.SetMaxIdleConns(500)
	_db.SetMaxOpenConns(500)
	_db.SetConnMaxLifetime(30 * time.Minute)
	return _db
}
