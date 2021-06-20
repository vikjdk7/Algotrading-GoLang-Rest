package controllers

import (
	"context"
	"fmt"
	"log"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/database"

	helper "github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/helpers"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var otpCollection *mongo.Collection = database.OpenCollection(database.Client, "otp")
var validate = validator.New()

//HashPassword is used to encrypt the password before it is stored in the DB
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

//VerifyPassword checks the input password while verifying it with the passward in the DB.
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or passowrd is incorrect")
		check = false
	}

	return check, msg
}

//CreateUser is the api used to get a single user
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exists"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}

//Login is the api used to get a single user
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or passowrd is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		c.JSON(http.StatusOK, foundUser)

	}
}

func ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var resetReq models.ResetPassword
		var foundUser models.User

		if err := c.BindJSON(&resetReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(resetReq)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": resetReq.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user with email not found"})
			return
		}
		passwordIsValid, msg := VerifyPassword(*resetReq.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		new_password := HashPassword(*resetReq.NewPassword)
		foundUser.Password = &new_password

		helper.UpdatePassword(foundUser.User_id, new_password)
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		c.JSON(http.StatusOK, foundUser)
	}
}

func ForgotPasswordEmailSend() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var forgotPasswordEmail models.ForgotPasswordEmail

		if err := c.BindJSON(&forgotPasswordEmail); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate.Struct(forgotPasswordEmail)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": forgotPasswordEmail.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}

		if count < 1 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Email. The email does not exist"})
			return
		}

		otp := helper.GenerateOtp()

		helper.SaveOtp(otp, *forgotPasswordEmail.Email)
		helper.SendEmail(otp, *forgotPasswordEmail.Email)

		response := models.ForgotPasswordEmailResponse{
			EmailSent: true,
		}
		c.JSON(http.StatusOK, response)
	}
}

func ForgotPasswordReset() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var forgotPasswordReset models.ForgotPasswordReset
		var foundUser models.User

		if err := c.BindJSON(&forgotPasswordReset); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(forgotPasswordReset)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": forgotPasswordReset.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or passowrd is incorrect"})
			return
		}

		var foundOtp models.Otp

		err = otpCollection.FindOne(ctx, bson.M{"email": foundUser.Email}).Decode(&foundOtp)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No Otp found for this email"})
			return
		}
		if foundOtp.Otp != *forgotPasswordReset.Otp {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid OTP"})
			return
		}
		if foundOtp.IsChecked == true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "OTP has already been used"})
			return
		}
		if time.Now().Sub(foundOtp.UpdatedAt).Minutes() > 5 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "OTP has expired"})
			return
		}

		new_password := HashPassword(*forgotPasswordReset.NewPassword)
		foundUser.Password = &new_password

		helper.UpdatePassword(foundUser.User_id, new_password)
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		_, err = userCollection.UpdateOne(
			ctx,
			bson.M{"email": forgotPasswordReset.Email},
			bson.D{
				{"$set", bson.M{"is_checked": true}},
			},
			&opt,
		)
		c.JSON(http.StatusOK, foundUser)

	}
}
