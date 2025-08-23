package database

import (
	"log"
	"raft-consensus/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "host=localhost user=postgres password=admin dbname=raft_consensus port=5433"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	err = DB.AutoMigrate(&models.RaftNode{})
	if err != nil {
		log.Fatal("Failed to automigrate: ", err)
	}

	log.Println("Database connected successfully!")
}