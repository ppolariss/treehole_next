package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"os"
	"treehole_next/config"
)

var DB *gorm.DB
var gormConfig = &gorm.Config{
	NamingStrategy: schema.NamingStrategy{
		SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
	},
}

func InitDB() {
	var err error
	if config.Config.Debug {
		err = os.MkdirAll("data", 0750)
		if err != nil {
			panic(err)
		}
		DB, err = gorm.Open(sqlite.Open("data/sqlite.db"), gormConfig)
	} else {
		DB, err = gorm.Open(mysql.Open(config.Config.DbUrl), gormConfig)
	}
	if err != nil {
		panic(err)
	}
	err = DB.AutoMigrate(
		&Division{},
		&Tag{},
		&Hole{},
		&Floor{},
		&FloorHistory{},
		&User{},
	)
	if err != nil {
		panic(err)
	}
}