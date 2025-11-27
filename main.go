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
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/xid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "recipes-api/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Recipe struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags" gorm:"serializer:json"`
	Ingredients  []string  `json:"ingredients" gorm:"serializer:json"`
	Instructions []string  `json:"instructions" gorm:"serializer:json"`
	PublishedAt  time.Time `json:"publishedAt"`
}

var db *gorm.DB

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

	if err := db.AutoMigrate(&Recipe{}); err != nil {
		log.Fatalf("Error migrating tables")
	}

	fmt.Println("Database connection established...")

	loadInitialData()
}

func loadInitialData() {
	file, err := os.ReadFile("recipes.json")
	if err != nil {
		log.Fatalf("Error reading recipes.json: %v", err)
	}

	var recipes []Recipe
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

	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)

	// swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":8080")
}

// @summary Create a recipe
// @Description Create a new recipe
// @Tags recipes
// @Accept json
// @Produce json
// @Param recipe body Recipe true "Recipe object"
// @Success 200 {object} Recipe
// @Router /recipes [post]
func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()

	if err := db.Create(&recipe).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recipe)
}

// @Summary List Recipes
// @Description Get all recipes
// @Tags recipes
// @Produce json
// @Success 200 {array} Recipe
// @Router /recipes [get]
func ListRecipesHandler(c *gin.Context) {
	var recipes []Recipe

	if err := db.Find(&recipes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipes"})
		return
	}

	c.JSON(http.StatusOK, recipes)
}

// @Summary Update an existing Recipe
// @Description Get an existing recipe and update it
// @Tags recipes
// @Accept json
// @produce json
// @Param id path string true "Recipe ID"
// @Param recipe body Recipe true "Recipe object"
// @Success 200 {object} Recipe
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /recipes/{id} [put]
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingRecipe Recipe
	if err := db.Where("id = ?", id).First(&existingRecipe).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}

	recipe.ID = existingRecipe.ID
	recipe.PublishedAt = existingRecipe.PublishedAt

	if err := db.Model(&existingRecipe).Updates(&recipe).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe"})
		return
	}

	c.JSON(http.StatusOK, existingRecipe)
}

// @Summary Delete a recipe
// @Description Delete a recipe by id
// @Tags recipes
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /recipes/{id} [delete]
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	var recipe Recipe
	if err := db.Where("id = ?", id).First(&recipe).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}

	if err := db.Delete(&recipe).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the recipe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been deleted"})
}

// @Summary Search recipes
// @Description Search recipes by tag
// @Tags recipes
// @Produce json
// @Param tag query string true "Tag to search for"
// @Success 200 {array} Recipe
// @Router /recipes/search [get]
func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag is required"})
	}

	var recipes []Recipe
	if err := db.Find(&recipes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search recipes"})
		return
	}

	var listOfRecipes []Recipe
	lowerTag := strings.ToLower(tag)

	for _, recipe := range recipes {
		for _, t := range recipe.Tags {
			if strings.Contains(strings.ToLower(t), lowerTag) {
				listOfRecipes = append(listOfRecipes, recipe)
			}
		}
	}

	c.JSON(http.StatusOK, listOfRecipes)
}
