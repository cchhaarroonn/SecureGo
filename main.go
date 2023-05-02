package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	uri string `json:"uri"`
}

func main() {
	//Enable weather mutiple accounts can be created with one license or only single account
	var onlyOneAccount bool = false

	//Get context and then connect to DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("YOUR DATABASE CONNECT URI"))
	if err != nil {
		errors.New("[X] Error occured while trying to connect to DB")
	}

	//Try to ping DB if it goes through we are inside it
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[*] Successfully connected to Database!")

	database := client.Database("securego")
	licenseCollection := database.Collection("licenses")
	usersCollection := database.Collection("users")

	server := gin.Default()

	//Create user account by username, password, license key
	server.POST("/securego/createUser/:username/:password/:license", func(c *gin.Context) {
		username := c.Param("username")
		password := c.Param("password")
		license := c.Param("license")

		doc := bson.D{
			{Key: "Username", Value: username},
			{Key: "Password", Value: password},
			{Key: "License", Value: license},
		}

		docLicense := bson.D{
			{Key: "License", Value: license},
		}

		var resultLic bson.D
		if err := licenseCollection.FindOne(context.Background(), docLicense).Decode(&resultLic); err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, gin.H{
					"status": "License doesn't exist in database",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to check license in database",
				"error":  err.Error(),
			})
			return
		}

		var resultAcc bson.D
		err := usersCollection.FindOne(context.Background(), doc).Decode(&resultAcc)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				if onlyOneAccount {
					// Check if an account with the same license already exists in the usersCollection
					docLicenseUser := bson.D{
						{Key: "License", Value: license},
					}
					var resultLicUser bson.D
					if err := usersCollection.FindOne(context.Background(), docLicenseUser).Decode(&resultLicUser); err == nil {
						c.JSON(http.StatusOK, gin.H{
							"status": "An account with this license already exists",
						})
						return
					}
				}
				_, err = usersCollection.InsertOne(context.Background(), doc)
				if err != nil {
					log.Fatal(err)
				}

				c.JSON(http.StatusCreated, gin.H{
					"status": "Account successfully created",
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status": "This account already exists",
				})
			}
		}
	})

	//Get info about user by his username
	server.GET("/securego/getUser/:name", func(c *gin.Context) {
		// Get the username parameter from the request URL
		username := c.Param("name")

		// Create a filter for the user with the specified username
		filter := bson.M{"Username": username}

		// Find the user in the database
		var user bson.M
		err := usersCollection.FindOne(context.Background(), filter).Decode(&user)
		if err != nil {
			// Return an error message if the user is not found
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"status": "User not found",
				})
				return
			}
			// Return an error message if there was a problem with the database
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Error fetching user from database",
				"error":  err.Error(),
			})
			return
		}

		// Extract the password and license key from the user document
		password := user["Password"].(string)
		license := user["License"].(string)

		// Return the password and license key as a JSON response
		c.JSON(http.StatusOK, gin.H{
			"username": username,
			"password": password,
			"license":  license,
		})
	})

	//Check if user exists by username
	server.POST("/securego/checkUser/:name", func(c *gin.Context) {
		name := c.Param("name")

		//Using this as our document we want to provide and also as filter for checking
		doc := bson.D{
			{Key: "Username", Value: name},
		}

		var result bson.D
		if err := usersCollection.FindOne(context.Background(), doc).Decode(&result); err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, gin.H{
					"status": "User doesn't exist in database",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to check user in database",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "User exists in database",
		})
	})

	//Remove user account by username and license
	server.POST("/securego/removeUser/:username/:license", func(c *gin.Context) {
		username := c.Param("username")
		license := c.Param("license")

		doc := bson.D{
			{Key: "Username", Value: username},
			{Key: "License", Value: license},
		}

		result, err := usersCollection.DeleteOne(context.Background(), doc)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to remove account from database",
				"error":  err.Error(),
			})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusOK, gin.H{
				"status": "This account doesn't exist in database",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "Account removed from database",
		})
	})

	//Remove user account by username
	server.POST("/securego/removeUser/:username", func(c *gin.Context) {
		username := c.Param("username")

		doc := bson.D{
			{Key: "Username", Value: username},
		}

		result, err := usersCollection.DeleteOne(context.Background(), doc)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to remove account from database",
				"error":  err.Error(),
			})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusOK, gin.H{
				"status": "This account doesn't exist in database",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "Account removed from database",
		})
	})

	//Create license key autotmatically with random letters
	server.POST("/securego/createLicense", func(c *gin.Context) {
		licenseKey := generateRandomString(10)

		doc := bson.D{
			{Key: "License", Value: licenseKey},
		}

		_, err = licenseCollection.InsertOne(context.Background(), doc)
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	})

	//Create specific licence key
	server.POST("/securego/createLicense/:name", func(c *gin.Context) {
		name := c.Param("name")

		//Using this as our document we want to provide and also as filter for checking
		doc := bson.D{
			{Key: "License", Value: name},
		}

		var result bson.D
		err := licenseCollection.FindOne(context.Background(), doc).Decode(&result)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				_, err = licenseCollection.InsertOne(context.Background(), doc)
				if err != nil {
					log.Fatal(err)
				}

				c.JSON(http.StatusOK, gin.H{
					"status": "success",
				})
			}
		}
	})

	//Remove license from the database, also remove all users who had that license
	server.POST("/securego/removeLicense/:name", func(c *gin.Context) {
		name := c.Param("name")

		doc := bson.D{
			{Key: "License", Value: name},
		}

		_, err := usersCollection.DeleteMany(context.Background(), bson.D{{Key: "License", Value: name}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to remove users with the license",
				"error":  err.Error(),
			})
			return
		}

		result, err := licenseCollection.DeleteOne(context.Background(), doc)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to remove license from database",
				"error":  err.Error(),
			})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusOK, gin.H{
				"status": "This license doesn't exist in database",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "License removed from database",
		})
	})

	//Check license key by the name that is specified
	server.GET("/securego/checkLicense/:name", func(c *gin.Context) {
		name := c.Param("name")

		doc := bson.D{
			{Key: "License", Value: name},
		}

		var result bson.D
		if err := licenseCollection.FindOne(context.Background(), doc).Decode(&result); err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, gin.H{
					"status": "License doesn't exist in database",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to check license in database",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "License exists in database",
		})
	})

	//Get all licenses in database
	server.GET("/securego/getLicenses/", func(c *gin.Context) {
		// Get all documents where value is not empty
		filter := bson.M{"License": bson.M{"$ne": ""}}
		cur, err := licenseCollection.Find(context.Background(), filter)
		if err != nil {
			log.Fatal(err)
		}

		// Loop through the results
		var results []bson.M
		if err := cur.All(context.Background(), &results); err != nil {
			log.Fatal(err)
		}
		licenses := make([]string, len(results))
		for i, result := range results {
			license := result["License"].(string)
			licenses[i] = license
		}

		// Return the list of licenses as a JSON response
		c.JSON(http.StatusOK, gin.H{
			"licenses": licenses,
		})
	})
	server.Run("localhost:42069")
}

// Helpers
func generateRandomString(length int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
