package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//User is the model that governs all notes objects retrived or inserted into the DB
type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	First_name    *string            `json:"first_name" validate:"required,min=2,max=100"`
	Last_name     *string            `json:"last_name" validate:"required,min=2,max=100"`
	Password      *string            `json:"Password" validate:"required,min=6""`
	Email         *string            `json:"email" validate:"email,required"`
	Phone         *string            `json:"phone" validate:"required"`
	Token         *string            `json:"token"`
	Refresh_token *string            `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
	User_id       string             `json:"user_id"`
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
	Phone      *string `json:"phone" validate:"required"`
	Otp        *string `json:"otp" validate:"required"`
}

type SignUpEmail struct {
	First_name *string `json:"first_name" validate:"required,min=2,max=100"`
	Last_name  *string `json:"last_name" validate:"required,min=2,max=100"`
	Email      *string `json:"email" validate:"email,required"`
}

type ForgotPasswordEmail struct {
	Email   *string `json:"email" validate:"email,required"`
	Process *string `json:"process" validate:"required,eq=forgot-password"`
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
