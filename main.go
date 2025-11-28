// @title Recipes API
// @version 1.0.0
// @description A recipes API server
// @contact.name Alex N. Kinuthia
// @contact.email alexkienjeku@gmail.com
// @host localhost:8080
// @BasePath /
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"

	"github.com/rs/xid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "recipes-api/docs"
	"recipes-api/handlers"
	"recipes-api/models"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var db *gorm.DB
var redisClient *redis.Client

func init() {
	var err error

	if err = godotenv.Load(); err != nil {
		log.Fatal("Failed to load environment variables")
	}

	host := os.Getenv("HOST")
	dbUser := os.Getenv("DBUSER")
	password := os.Getenv("PASSWORD")
	dbName := os.Getenv("DBNAME")
	port := os.Getenv("PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Africa/Nairobi", host, dbUser, password, dbName, port)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	if err := db.AutoMigrate(&models.Recipe{}); err != nil {
		log.Fatalf("Error migrating tables")
	}

	fmt.Println("Database connection established...")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping()
	fmt.Println(status)

	loadInitialData()
}

func loadInitialData() {
	file, err := os.ReadFile("recipes.json")
	if err != nil {
		log.Fatalf("Error reading recipes.json: %v", err)
	}

	var recipes []models.Recipe
	if err := json.Unmarshal(file, &recipes); err != nil {
		log.Fatalf("Error parsing recipes.json: %v", err)
	}

	if err := db.Exec("DELETE FROM recipes").Error; err != nil {
		log.Fatalf("Error clearing recipes table: %v", err)
	}

	for _, recipe := range recipes {
		if recipe.ID == "" {
			recipe.ID = xid.New().String()
		}
		if recipe.PublishedAt.IsZero() {
			recipe.PublishedAt = time.Now()
		}

		if err := db.Create(&recipe).Error; err != nil {
			log.Fatalf("Error inserting recipe %s: %v", recipe.Name, err)
		}
	}

	log.Printf("Successfully loaded %d recipes from recipes.json into database", len(recipes))
}

func main() {
	router := gin.Default()

	rh := handlers.NewRecipeController(db, redisClient)

	router.POST("/recipes", rh.NewRecipeHandler)
	router.GET("/recipes", rh.ListRecipesHandler)
	router.PUT("/recipes/:id", rh.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", rh.DeleteRecipeHandler)
	router.GET("/recipes/search", rh.SearchRecipesHandler)

	// swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":8080")
}
