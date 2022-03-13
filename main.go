// Recipes API
//
// This is a sample recipes API.
//
// 	Schemes: http
// 	Host: localhost:8080
// 	BasePath: /
// 	Version: 1.0.0
// 	Contact: Javi<hi@orov.io>
// 	Consumes:
// 		- application/json
//
// 	Produces:
// 		- application/json
// swagger:meta
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// swagger:parameters recipes newRecipe
type Recipe struct {
	// swagger:ignore
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

var recipes []Recipe
var ctx context.Context
var collection *mongo.Collection

func init() {
	ctx = context.Background()
	mongoURI := os.Getenv("MONGO_URI")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Unable to connect to mongo DB due to " + err.Error())
	}
	if err = mongoClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("Unable to ping to mongo DB due to " + err.Error())
	}
	log.Println("Connected to MongoDB")

	// Move this to a seed script with a manager...
	// recipes = make([]Recipe, 0)
	// file, _ := ioutil.ReadFile("recipes.json")
	// _ = json.Unmarshal([]byte(file), &recipes)
	// var listOfRecipes []interface{}
	// for _, recipe := range recipes {
	// 	listOfRecipes = append(listOfRecipes, recipe)
	// }

	collection = mongoClient.Database(os.Getenv(("MONGO_DATABASE"))).Collection("recipe")
	// insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	// if err != nil {
	// 	log.Fatal("Unable to seed database due to " + err.Error())
	// }

	// log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))
}

// swagger:operation POST /recipes recipes newRecipe
// Create a new recipe
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '400':
//         description: Invalid input
func NewRecipe(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := collection.InsertOne(ctx, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorHuman": "Error while inserting a new recipe",
			"error":      err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful operation
func ListRecipes(c *gin.Context) {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer cursor.Close(ctx)
	recipes := make([]Recipe, 0)
	for cursor.Next(ctx) {
		var recipe Recipe
		cursor.Decode(&recipe)
		recipes = append(recipes, recipe)
	}

	c.JSON(http.StatusOK, recipes)

}

// swagger:operation PUT /recipes/{id} recipes updateRecipe
// Update an existing recipe
// ---
// parameters:
// - name: id
//   in: path
//   description: Id of the recipe
//   required: true
//   type: string
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful operation
//   '400':
//     description: Invalid input
//   '200':
//     description: Invalid recipe ID
func UpdateRecipe(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectID, _ := primitive.ObjectIDFromHex(id)
	result, err := collection.UpdateOne(ctx, bson.M{
		"_id": objectID,
	}, bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: recipe.Name},
		{Key: "instructions", Value: recipe.Instructions},
		{Key: "ingredients", Value: recipe.Ingredients},
		{Key: "tags", Value: recipe.Tags},
	}}})

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "recipe not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorHuman": "Error while  recipe",
			"error":      err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been updated",
	})
}

// swagger:operation DELETE /recipes/{id} recipes deleteRecipe
// Delete an existing recipe
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
// responses:
//     '200':
//         description: Successful operation
//     '404':
//         description: Invalid recipe ID
func DeleteRecipe(c *gin.Context) {
	id := c.Param("id")
	objectID, _ := primitive.ObjectIDFromHex(id)

	_, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorHuman": "Error while deleting a recipe",
			"error":      err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
	})
}

// swagger:operation GET /recipes/search recipes findRecipe
// Search recipes based on tags
// ---
// produces:
// - application/json
// parameters:
//   - name: tag
//     in: query
//     description: recipe tag
//     required: true
//     type: string
// responses:
//     '200':
//         description: Successful operation
func SearchRecipes(c *gin.Context) {
	tag := c.Query("tag")
	matchedRecipes := make([]Recipe, 0)
	for i := 0; i < len(recipes); i++ {
		for _, t := range recipes[i].Tags {
			if strings.EqualFold(t, tag) {
				matchedRecipes = append(matchedRecipes, recipes[i])
				break
			}
		}
	}
	c.JSON(http.StatusOK, matchedRecipes)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipe)
	router.GET("/recipes", ListRecipes)
	router.PUT("/recipes/:id", UpdateRecipe)
	router.DELETE("/recipes/:id", DeleteRecipe)
	router.Run()
}
