package helpers

import (
	"context"
	"errors"
	"go-jwt-project/database"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	Uid        string
	User_Type  string
	jwt.StandardClaims
}

var user_collection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokes(
	email,
	first_name,
	last_name,
	user_type,
	user_id string,
) (
	signed_token,
	signed_refresh_token string,
	error error,
) {
	claims := &SignedDetails{
		Email:      email,
		First_Name: first_name,
		Last_Name:  last_name,
		Uid:        user_id,
		User_Type:  user_type,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(1)).Unix(),
		},
	}

	refresh_claims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(1)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}
	refresh_token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refresh_claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	return token, refresh_token, err
}

func ValidateToken(signed_token string) (*SignedDetails, error) {
	token, err := jwt.ParseWithClaims(
		signed_token,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		return nil, errors.New("the token is invalid")
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, errors.New("the token is expired")
	}
	return claims, nil
}

func UpdateAllTokens(token, refresh_token, user_id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var update_object primitive.D

	update_object = append(update_object, bson.E{Key: "token", Value: token})
	update_object = append(update_object, bson.E{Key: "refresh_token", Value: refresh_token})

	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	update_object = append(update_object, bson.E{Key: "updated_at", Value: updated_at})

	upsert := true
	filter := bson.M{"user_id": user_id}
	opts := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := user_collection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: update_object}}, &opts)
	if err != nil {
		log.Panic(err)
		return
	}
}
