package models

import (
	"time"

	stripe "github.com/stripe/stripe-go/v72"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ErrorString struct {
	S string
}

func (e *ErrorString) Error() string {
	return e.S
}

type Plan struct {
	Active          bool                       `json:"active"`
	AggregateUsage  string                     `json:"aggregate_usage"`
	Amount          int64                      `json:"amount"`
	AmountDecimal   float64                    `json:"amount_decimal,string"`
	BillingScheme   stripe.PlanBillingScheme   `json:"billing_scheme"`
	Created         int64                      `json:"created"`
	Currency        stripe.Currency            `json:"currency"`
	Deleted         bool                       `json:"deleted"`
	ID              string                     `json:"id"`
	Interval        stripe.PlanInterval        `json:"interval"`
	IntervalCount   int64                      `json:"interval_count"`
	Livemode        bool                       `json:"livemode"`
	Nickname        string                     `json:"nickname"`
	Product         *stripe.Product            `json:"product"`
	TransformUsage  *stripe.PlanTransformUsage `json:"transform_usage"`
	TrialPeriodDays int64                      `json:"trial_period_days"`
	UsageType       stripe.PlanUsageType       `json:"usage_type"`
}

/*type UserSubscription struct {
	ID                   primitive.ObjectID `bson:"_id"`
	User_id              string             `json:"user_id" bson:"user_id"`
	Email                string             `json:"email" bson:"email"`
	StripeCustomerId     string             `json:"stripe_customer_id" bson:"stripe_customer_id"`
	StripeSubscriptionId string             `json:"stripe_subscription_id" bson:"stripe_subscription_id"`
}*/
type User struct {
	ID                primitive.ObjectID `bson:"_id"`
	First_name        *string            `json:"first_name" validate:"required,min=2,max=100"`
	Last_name         *string            `json:"last_name" validate:"required,min=2,max=100"`
	Password          *string            `json:"Password" validate:"required,min=6""`
	Email             *string            `json:"email" validate:"email,required"`
	Token             *string            `json:"token"`
	Refresh_token     *string            `json:"refresh_token"`
	Created_at        time.Time          `json:"created_at"`
	Updated_at        time.Time          `json:"updated_at"`
	User_id           string             `json:"user_id"`
	Stripe_customerId string             `json:"stripe_customer_id`
}
