package controllers

import (
	"context"
	"fmt"
	"go-jwt-project/database"
	"go-jwt-project/helpers"
	"go-jwt-project/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var user_collection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Panic(err)
	}
	return string(hashedPassword)
}

func VerifyPassword(user_password, provided_password string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(user_password), []byte(provided_password))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("error occurred while comparing the password %v", err)
		check = false
	}
	return check, msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validation_err := validate.Struct(user)
		if validation_err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validation_err.Error()})
			return
		}

		count, err := user_collection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking for the email"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email already exists"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		count, err = user_collection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking for the phone number"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "phone number already exists"})
			return
		}

		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_Id = user.ID.Hex()
		token, refresh_token, _ := helpers.GenerateAllTokes(*user.Email, *user.First_Name, *user.Last_Name, *user.User_Type, user.User_Id)
		user.Token = &token
		user.Refresh_Token = &refresh_token

		result_insertion_number, insert_err := user_collection.InsertOne(ctx, user)
		if insert_err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User item was not created"})
			return
		}
		c.JSON(http.StatusOK, result_insertion_number)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var found_user models.User
		err := user_collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&found_user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		password_is_valid, msg := VerifyPassword(*user.Password, *found_user.Password)
		if !password_is_valid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if found_user.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		token, refresh_token, _ := helpers.GenerateAllTokes(*found_user.Email, *found_user.First_Name, *found_user.Last_Name, *found_user.User_Type, found_user.User_Id)
		helpers.UpdateAllTokens(token, refresh_token, found_user.User_Id)
		err = user_collection.FindOne(ctx, bson.M{"user_id": found_user.User_Id}).Decode(&found_user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, found_user)
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Param("id")
		if err := helpers.MatchUserTypeToUid(c, user_id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := user_collection.FindOne(ctx, bson.M{"user_id": user_id}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar el tipo de usuario
		if err := helpers.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Crear un contexto con timeout
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Obtener y validar parámetros de la consulta
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		if startIndexParam := c.Query("startIndex"); startIndexParam != "" {
			startIndex, err = strconv.Atoi(startIndexParam)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid startIndex parameter"})
				return
			}
		}

		// Definir las etapas de la agregación
		matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
		groupStage := bson.D{
			{
				Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
					{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
				},
			},
		}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{
					{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
				}},
			}},
		}

		// Ejecutar la agregación
		result, err := user_collection.Aggregate(ctx, mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Procesar los resultados
		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}

		// Manejar el caso donde no se encuentren usuarios
		if len(allUsers) == 0 {
			c.JSON(http.StatusOK, gin.H{"total_count": 0, "user_items": []bson.M{}})
			return
		}

		c.JSON(http.StatusOK, allUsers[0])
	}
}
