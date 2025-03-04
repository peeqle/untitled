package service

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
	"regexp"
	"untitled/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func FetchUserByUsername(username string) (models.User, error) {
	var result models.User

	err := DB.Where("username = ?", username).First(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, errors.New("user not found")
	}

	if err != nil {
		return models.User{}, err
	}

	return result, nil
}

func ConnectDB() {
	dsn := os.Getenv("DATABASE_URL")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Trying to create db and connect", err)

		tryCreateDatabase()
	}

	if err := DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations applied successfully")
}

func tryCreateDatabase() {
	cp := collectConnectionParticles()
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d?sslmode=disable", cp.username, cp.password, cp.host, cp.port)

	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s;", cp.databaseName))
	if err != nil {
		log.Fatalf("Failed to create database: %v\n", err)
	}

	fmt.Printf("Database %s created successfully!\n", cp.databaseName)
	ConnectDB()
}

func collectConnectionParticles() ConnectionParticles {
	dsn := os.Getenv("DATABASE_URL")
	cp := ConnectionParticles{}

	//username
	r, _ := regexp.Compile("([a-zA-Z0-9_-]+)")
	cp.username = r.FindStringSubmatch(dsn)[1]

	//password
	r, _ = regexp.Compile(":([a-zA-Z0-9_-]+)@")
	cp.password = r.FindStringSubmatch(dsn)[1]

	//host
	r, _ = regexp.Compile("@([a-zA-Z0-9.-]+)")
	cp.host = r.FindStringSubmatch(dsn)[1]

	//port
	r, _ = regexp.Compile(":(\\d+)")
	cp.port = r.FindStringSubmatch(dsn)[1]

	//dbName
	r, _ = regexp.Compile("\\/([a-zA-Z0-9_-]+)")
	cp.databaseName = r.FindStringSubmatch(dsn)[1]

	//sslEnabled
	r, _ = regexp.Compile("(\\?sslEnabled=(true|false))?$")
	cp.sslEnabled = r.FindStringSubmatch(dsn)[1] == "true"

	return cp
}

type ConnectionParticles struct {
	username     string
	password     string
	host         string
	port         string
	databaseName string
	sslEnabled   bool
}
