package util

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // mysql TODO
	"github.com/spf13/viper"
)

// Database TODO
type Database struct {
	Self *gorm.DB
}

// DB TODO
var DB *Database

// setupDatabase initialize the database tables.
func setupDatabase(db *gorm.DB) {
}

func openDB(username, password, addr, name string) *gorm.DB {
	config := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=%t&loc=%s",
		username,
		password,
		addr,
		name,
		true,
		// "Asia/Shanghai"),
		"Local")
	db, err := gorm.Open("mysql", config)
	if err != nil {
		log.Fatalf("Database connection failed. Database name: %s, error: %v", name, err)
	}

	// set for db connection
	setupDB(db)
	return db
}

func setupDB(db *gorm.DB) {
	// setup tables
	setupDatabase(db)

	db.LogMode(viper.GetBool("gormlog"))
	db.DB().SetMaxIdleConns(60) // 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	db.DB().SetMaxOpenConns(200)
	/*
		First of all, you should use DB.SetConnMaxLifetime() instead of wait_timeout.
		Closing connection from client is always better than closing from server,
		because client may send query just when server start closing the connection.
		In such case, client can't know sent query is received or not.
	*/
	db.DB().SetConnMaxLifetime(3600 * time.Second)
}

// initSelfDB TODO
// used for cli
func initSelfDB() *gorm.DB {
	return openDB(viper.GetString("db.user"),
		viper.GetString("db.password"),
		fmt.Sprintf("%s:%s", viper.GetString("db.host"), viper.GetString("db.port")),
		viper.GetString("db.name"))
}

// Init TODO
func (db *Database) Init() {
	DB = &Database{
		Self: initSelfDB(),
	}
}

// Close TODO
func (db *Database) Close() {
	DB.Self.Close()
}
