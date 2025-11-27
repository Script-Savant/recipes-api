package handlers

import (
	"net/http"
	"recipes-api/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

type RecipeController struct {
	db *gorm.DB
}

func NewRecipeController(db *gorm.DB) *RecipeController {
	return &RecipeController{db}
}

// @summary Create a recipe
// @Description Create a new recipe
// @Tags recipes
// @Accept json
// @Produce json
// @Param recipe body Recipe true "Recipe object"
// @Success 200 {object} Recipe
// @Router /recipes [post]
func (r *RecipeController) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()

	if err := r.db.Create(&recipe).Error; err != nil {
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
func (r *RecipeController) ListRecipesHandler(c *gin.Context) {
	var recipes []models.Recipe

	if err := r.db.Find(&recipes).Error; err != nil {
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
func (r *RecipeController) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingRecipe models.Recipe
	if err := r.db.Where("id = ?", id).First(&existingRecipe).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}

	recipe.ID = existingRecipe.ID
	recipe.PublishedAt = existingRecipe.PublishedAt

	if err := r.db.Model(&existingRecipe).Updates(&recipe).Error; err != nil {
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
func (r *RecipeController) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	var recipe models.Recipe
	if err := r.db.Where("id = ?", id).First(&recipe).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}

	if err := r.db.Delete(&recipe).Error; err != nil {
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
func (r *RecipeController) SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag is required"})
	}

	var recipes []models.Recipe
	if err := r.db.Find(&recipes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search recipes"})
		return
	}

	var listOfRecipes []models.Recipe
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
