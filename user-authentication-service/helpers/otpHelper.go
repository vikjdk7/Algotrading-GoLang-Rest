package helper

import (
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"time"

	"github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var otpCollection *mongo.Collection = database.OpenCollection(database.Client, "otp")

func GenerateOtp() int64 {
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatal(err)
	}
	return n.Int64()
}

func SaveOtp(otp string, email string, process string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj primitive.D
	updateObj = append(updateObj, bson.E{"email", email})
	updateObj = append(updateObj, bson.E{"otp", otp})
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", updatedAt})
	updateObj = append(updateObj, bson.E{"is_checked", false})
	updateObj = append(updateObj, bson.E{"process", process})

	upsert := true
	filter := bson.M{"email": email, "process": process}

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
