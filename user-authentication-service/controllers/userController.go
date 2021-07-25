package controllers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/database"

	helper "github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/helpers"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/middleware"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/user-authentication-service/models"

	"cloud.google.com/go/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	gcsoptions "google.golang.org/api/option"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var otpCollection *mongo.Collection = database.OpenCollection(database.Client, "otp")
var userProfileCollection *mongo.Collection = database.OpenCollection(database.Client, "user_profile")
var validate = validator.New()
var uploader *models.ClientUploader

const (
	projectID  = "hedgina"
	bucketName = "hedgina-algobot"
)

func init() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", `{"type":"service_account","project_id":"hedgina","private_key_id":"c7e362ab4968ea8f56f8d3db6670664e01c0ec65","private_key":"-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDTYkgoLnNR7GBQ\niFmG3X63eo/yzy6Mm3gJGG2GzOHm4U+5KFuaINNa81UvB+WJCM8S7AZ2ptDy6lBc\nPHu+bU7cP/VIU00kyn+8+K59sfg3E7bbH+5vr9SoI38fwwF0Y/F21lYWzuiJ0wm7\ns4EVuZUIXBDdb3NCXTPCjxMOmWywewg/nItUFpJXCsQal7dLcAR/jnfn21xUqK42\nrigxa0YCAxnZDIrrXDYDLJIMnXF5vQbDl45l5NvcmhQ2BybS6RozNjKdgkMyC87q\ncwZif1OkHVtspvFnwQNbK6jEsxCDjhApvjIbNfXqU4YW7fPm9ZWmROvvnVueDIZ1\nZwFy1BZrAgMBAAECggEAGUPA7hSHMf53kIiLcsQcdh+O/u1mWeXnFec2iAsK4QaC\n+dVqBWTw/gjhYIqoE5Xa8h8Fsr7DcJUy36NXAu6bo1V9opRaPxB47gQnDtzrluGj\nVFNfszenyiTb99bd4KRlYtfBWF6IipiPrECLKCaTOnmOhnjgpMjw+8sP8wnBZOJH\n8DpPfVQcS6Amq2RTYN9NWfjfFSzhQz0tHUMa2fV9ymyhKRQCfjTmAeNjuGm0qgsM\n3p4zJPbW2KFufv9zLJBEO2PKfEzi2N1nHAbPkiWOa3mwT/CuCpKhIMt8Ws2Whkio\nRGABiJL97FPYzy3B/Z2+sLmi6b4nuEUzshrIdg7W4QKBgQD0Y9SGGcv9CXOFxPLj\nTWN/G4EgQOl64MRp1T0LrLNoTxyasVTwsht9itK0+s1ozBiqBzSIzig6gU8i9ro1\nWwCF0YMZ53wLmsFDIxB73RRDrs71vOrg6t5HyIUcaPHjFAO+f9j/Ytb38Dqih8Lx\ncFDxuyir1DBSBu1XLQrKmLJisQKBgQDdbQu27s71ZcLU+2ArkwZm9Gb87Lj/z4Sx\nVDLIXAzhdcVWea2/FQcxXmNraD5EwackoZ4m255OsBgVrIbApAPgP5QeKn3crVfy\nUlo4ZF9PluqXs9H2TRUSuPd/TwsHVRgssvB5+zBnVYRX+Pl9kLMTajNvF4B34QTY\nCzIy99R52wKBgAZJgzAn/b12vsgUNwNt/D9K39mKkfcdTTBD0hw4xyzJzDyWj07Z\n5icmqSEKyroFdiT5pnpWg2Zt6TFHE6dHvg2zRCIoeGJ8CrjFcCkfmOPc3WopAAnl\nQO6r0/DVKlPjMe12sIhxbIJYZcnEoFlBwHNXk0ZIYS3bC8QQXpSztPMhAoGANGLF\nH712B0bRBnSGdyisnhT6fKJAzny5Jv8FmLN2dKzZSDE3cvq1ne931ARwnvG16ou2\nD/lrhbBRsmcD5nWnWRmRoGVrK5dzNChZoffVOM46qDNp3Dy2XJyYKW147X4rXv/i\ntuk/tWLdEbccx6FBTLmWe5Ty1unMrJRRhw9tHHsCgYEAhXLwNm2Xtkmzb8Q9Mffk\nPlEZbEZYkZHwIVV3PvNhIs3v9uM2KN2dAhLGvJDyWePKKWNiMWhcpSIZv+i3ybz9\n9E1GsZWUGGdSONNX2uVGFlm3JxU3ozk8Olh6C4w+k+5TgUEUpM2no2qXsUkMMSLs\n6YUuiXOlYWeiYmn3txkGukE=\n-----END PRIVATE KEY-----\n","client_email":"gcp-image-admin@hedgina.iam.gserviceaccount.com","client_id":"109441199199133427068","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_x509_cert_url":"https://www.googleapis.com/robot/v1/metadata/x509/gcp-image-admin%40hedgina.iam.gserviceaccount.com"}`)
	GOOGLE_APPLICATION_CREDENTIALS := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	client, err := storage.NewClient(context.Background(), gcsoptions.WithCredentialsJSON([]byte(GOOGLE_APPLICATION_CREDENTIALS)))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	uploader = &models.ClientUploader{
		Cl:         client,
		BucketName: bucketName,
		ProjectID:  projectID,
		UploadPath: "profile-images/",
	}

}

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

func SignUpEmailSend() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var signUpEmail models.SignUpEmail

		if err := c.BindJSON(&signUpEmail); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate.Struct(signUpEmail)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": signUpEmail.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An account with this email already exists"})
			return
		}

		otp := strconv.FormatInt(helper.GenerateOtp(), 10)

		helper.SaveOtp(otp, *signUpEmail.Email, "registration")
		helper.SendSignUpEmail(otp, *signUpEmail.Email, *signUpEmail.First_name, *signUpEmail.Last_name)

		response := models.EmailResponse{
			EmailSent: true,
		}
		c.JSON(http.StatusOK, response)
	}
}

//CreateUser is the api used to get a single user
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var userSignUp models.UserSignUp
		var user models.User

		if err := c.BindJSON(&userSignUp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(userSignUp)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": userSignUp.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}

		password := HashPassword(*userSignUp.Password)
		user.Password = &password

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email already exists"})
			return
		}
		var foundOtp models.Otp
		err = otpCollection.FindOne(ctx, bson.M{"email": userSignUp.Email, "process": "registration"}).Decode(&foundOtp)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No Otp found for this email"})
			return
		}

		if foundOtp.Otp != *userSignUp.Otp {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid OTP"})
			return
		}
		if foundOtp.IsChecked == true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "OTP has already been used"})
			return
		}
		if time.Now().Sub(foundOtp.UpdatedAt).Minutes() > 10 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "OTP has expired"})
			return
		}

		user.First_name = userSignUp.First_name
		user.Last_name = userSignUp.Last_name
		user.Email = userSignUp.Email
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

		var userProfile models.UserProfile
		userProfile.ID = primitive.NewObjectID()
		userProfile.User_id = &user.User_id
		userProfile.First_name = user.First_name
		userProfile.Last_name = user.Last_name
		receiveNot := true
		userProfile.ReceiveNotification = &receiveNot
		userProfile.Email = user.Email
		_, insertProfileErr := userProfileCollection.InsertOne(ctx, userProfile)
		if insertProfileErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		_, err = otpCollection.UpdateOne(
			ctx,
			bson.M{"email": userSignUp.Email, "process": "registration"},
			bson.M{
				"$set": bson.M{"is_checked": true},
			},
			&opt,
		)

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

		currentTimestamp, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id, currentTimestamp)
		foundUser.Token = &token
		foundUser.Refresh_token = &refreshToken
		foundUser.Updated_at = currentTimestamp

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
		currentTimestamp, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id, currentTimestamp)
		foundUser.Updated_at = currentTimestamp
		foundUser.Token = &token
		foundUser.Refresh_token = &refreshToken
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

		otp := strconv.FormatInt(helper.GenerateOtp(), 10)

		helper.SaveOtp(otp, *forgotPasswordEmail.Email, "forgot-password")
		helper.SendForgotPasswordEmail(otp, *forgotPasswordEmail.Email)

		response := models.EmailResponse{
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "incorrect email"})
			return
		}

		var foundOtp models.Otp

		err = otpCollection.FindOne(ctx, bson.M{"email": forgotPasswordReset.Email, "process": "forgot-password"}).Decode(&foundOtp)
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
		currentTimestamp, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id, currentTimestamp)

		foundUser.Updated_at = currentTimestamp
		foundUser.Token = &token
		foundUser.Refresh_token = &refreshToken
		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		_, err = otpCollection.UpdateOne(
			ctx,
			bson.M{"email": forgotPasswordReset.Email, "process": "forgot-password"},
			bson.M{
				"$set": bson.M{"is_checked": true},
			},
			&opt,
		)
		c.JSON(http.StatusOK, foundUser)

	}
}

func UserProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		//var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var userProfile models.UserProfile
		var updateUserDB bool

		if err := c.BindJSON(&userProfile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(userProfile)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		if len(c.Request.Header["Token"]) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token cannot be empty"})
			return
		}
		token := c.Request.Header["Token"][0]

		userId, errorMsg := middleware.ValdateIncomingToken(token)
		if errorMsg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}

		filter := bson.M{"user_id": userId}
		updateUser := bson.M{}
		updateProfile := bson.M{}

		if *userProfile.First_name != "" {
			updateUserDB = true
			updateUser["first_name"] = userProfile.First_name
			updateProfile["first_name"] = userProfile.First_name
		}
		if *userProfile.Last_name != "" {
			updateUserDB = true
			updateUser["last_name"] = userProfile.Last_name
			updateProfile["last_name"] = userProfile.Last_name
		}
		if userProfile.Email != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email cannot be changed"})
			return
		}
		if *userProfile.CurrentAddress != "" {
			updateProfile["current_address"] = userProfile.CurrentAddress
		}
		if *userProfile.BillingAddress != "" {
			updateProfile["billing_address"] = userProfile.BillingAddress
		}
		if userProfile.SameBillingAdd != nil {
			updateProfile["same_billing_add"] = userProfile.SameBillingAdd
			if *userProfile.SameBillingAdd == true {
				updateProfile["billing_address"] = userProfile.CurrentAddress
			}
		}
		if userProfile.ReceiveNotification != nil {
			updateProfile["receive_notification"] = userProfile.ReceiveNotification
		}
		/*if *userProfile.ProfilePicture != "" {
			updateProfile["profile_picture"] = userProfile.ProfilePicture
		}*/
		if updateUserDB == true {
			result := userCollection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": updateUser})

			// Decode result and write it to 'decoded'
			var decoded models.User
			err := result.Decode(&decoded)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "User cannot be updated"})
				return
			}

		}
		result := userProfileCollection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": updateProfile}, options.FindOneAndUpdate().SetReturnDocument(1))

		// Decode result and write it to 'decoded'
		var decodedProfile models.UserProfile
		err := result.Decode(&decodedProfile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User Profile cannot be updated"})
			return
		}
		c.JSON(http.StatusOK, decodedProfile)
	}
}

func GetUserProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		//var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var userProfile models.UserProfile

		if len(c.Request.Header["Token"]) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token cannot be empty"})
			return
		}
		token := c.Request.Header["Token"][0]

		userId, errorMsg := middleware.ValdateIncomingToken(token)
		if errorMsg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}

		err := userProfileCollection.FindOne(context.TODO(), bson.M{"user_id": userId}).Decode(&userProfile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find User Profile details for the user"})
			return
		}
		c.JSON(http.StatusOK, userProfile)
	}
}

func DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		//var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		//fmt.Println(c.Request)
		if len(c.Request.Header["Token"]) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token cannot be empty"})
			return
		}
		token := c.Request.Header["Token"][0]

		userId, errorMsg := middleware.ValdateIncomingToken(token)
		if errorMsg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}

		_, err := userProfileCollection.DeleteOne(context.TODO(), bson.M{"user_id": userId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete UserProfile"})
			return
		}

		deleteResult, err := userCollection.DeleteOne(context.TODO(), bson.M{"user_id": userId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete User"})
			return
		}

		if deleteResult.DeletedCount == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User Not Found"})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}

func UploadProfileImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.Header["Token"]) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token cannot be empty"})
			return
		}
		token := c.Request.Header["Token"][0]

		userId, errorMsg := middleware.ValdateIncomingToken(token)
		if errorMsg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}

		f, err := c.FormFile("profile_image")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		blobFile, err := f.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		//filename := fmt.Sprintf("%s_%s", userId, f.Filename)
		err = uploader.UploadFile(blobFile, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "success"})
	}
}

func GetProfileImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.Header["Token"]) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token cannot be empty"})
			return
		}
		token := c.Request.Header["Token"][0]

		userId, errorMsg := middleware.ValdateIncomingToken(token)
		if errorMsg != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}

		data, contentType, err := uploader.ReadFile(userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Data(http.StatusOK, contentType, data)
	}
}
