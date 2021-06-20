package helper

import (
	"context"
	"crypto/rand"
	"log"
	"time"

	"github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var otpCollection *mongo.Collection = database.OpenCollection(database.Client, "otp")

func GenerateOtp() string {
	p, _ := rand.Prime(rand.Reader, 20)
	return p.String()
}

func SaveOtp(otp string, email string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj primitive.D
	updateObj = append(updateObj, bson.E{"email", email})
	updateObj = append(updateObj, bson.E{"otp", otp})
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", updatedAt})
	updateObj = append(updateObj, bson.E{"is_checked", false})

	upsert := true
	filter := bson.M{"email": email}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := otpCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)

	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}

	return

}
