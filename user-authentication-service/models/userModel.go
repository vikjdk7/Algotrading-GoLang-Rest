package models

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"time"

	"cloud.google.com/go/storage"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//User is the model that governs all notes objects retrived or inserted into the DB
type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	First_name    *string            `json:"first_name" validate:"required,min=2,max=100"`
	Last_name     *string            `json:"last_name" validate:"required,min=2,max=100"`
	Password      *string            `json:"Password" validate:"required,min=6""`
	Email         *string            `json:"email" validate:"email,required"`
	Token         *string            `json:"token"`
	Refresh_token *string            `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
	User_id       string             `json:"user_id"`
}

type UserProfile struct {
	ID                  primitive.ObjectID `bson:"_id"`
	User_id             *string            `json:"user_id" bson:"user_id"`
	First_name          *string            `json:"first_name" validate:"min=2,max=100" bson:"first_name"`
	Last_name           *string            `json:"last_name" validate:"min=2,max=100" bson:"last_name"`
	Email               *string            `json:"email" bson:"email"`
	CurrentAddress      *string            `json:"current_address" bson:"current_address"`
	BillingAddress      *string            `json:"billing_address" bson:"billing_address"`
	SameBillingAdd      *bool              `json:"same_billing_add" bson:"same_billing_add"`
	ReceiveNotification *bool              `json:"receive_notification" bson:"receive_notification"`
	//ProfilePicture      *string            `json:"profile_picture" bson:"profile_picture"`
}

type ResetPassword struct {
	Email       *string `json:"email" validate:"email,required"`
	Password    *string `json:"Password" validate:"required,min=6"`
	NewPassword *string `json:"new_password" validate:"required,min=6"`
}

type UserSignUp struct {
	First_name *string `json:"first_name" validate:"required,min=2,max=100"`
	Last_name  *string `json:"last_name" validate:"required,min=2,max=100"`
	Password   *string `json:"Password" validate:"required,min=6""`
	Email      *string `json:"email" validate:"email,required"`
	Otp        *string `json:"otp" validate:"required"`
}

type SignUpEmail struct {
	First_name *string `json:"first_name" validate:"required,min=2,max=100"`
	Last_name  *string `json:"last_name" validate:"required,min=2,max=100"`
	Email      *string `json:"email" validate:"email,required"`
}

type ForgotPasswordEmail struct {
	Email *string `json:"email" validate:"email,required"`
}

type EmailResponse struct {
	EmailSent bool `json:"email_sent"`
}

type ForgotPasswordReset struct {
	Email       *string `json:"email" validate:"email,required"`
	Otp         *string `json:"otp" validate:"required"`
	NewPassword *string `json:"new_password" validate:"required,min=6"`
}

type Otp struct {
	Id        primitive.ObjectID `bson:"_id"`
	Email     string             `json:"email" bson:"email"`
	IsChecked bool               `json:"is_checked" bson:"is_checked"`
	Otp       string             `json:"otp" bson:"otp"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	Process   string             `json:"process" bson:"process"`
}

type ClientUploader struct {
	Cl         *storage.Client
	ProjectID  string
	BucketName string
	UploadPath string
}

// UploadFile uploads an object
func (c *ClientUploader) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := c.Cl.Bucket(c.BucketName).Object(c.UploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

func (c *ClientUploader) ReadFile(object string) ([]byte, string, error) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := c.Cl.Bucket(c.BucketName).Object(c.UploadPath + object).NewReader(ctx)
	if err != nil {
		return nil, "", err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, "", err
	}
	return data, rc.ContentType(), nil
}
