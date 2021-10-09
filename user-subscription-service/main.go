package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/user-subscription-service/helper"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/user-subscription-service/middleware"
	"github.com/vikjdk7/Algotrading-GoLang-Rest/user-subscription-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	stripe "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/plan"
	"github.com/stripe/stripe-go/v72/product"
)

func getPlans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var customError models.ErrorString

	token := r.Header.Get("token")

	if token == "" {
		customError.S = "Token cannot be empty"
		helper.GetError(&customError, w)
		return
	}

	_, _, _, _, errorMsg := middleware.ValdateIncomingToken(token)
	if errorMsg != "" {
		customError.S = errorMsg
		helper.GetError(&customError, w)
		return
	}

	params := &stripe.PlanListParams{}
	params.Filters.AddFilter("limit", "", "25")
	i := plan.List(params)
	var plans []models.Plan
	for i.Next() {
		var plan models.Plan
		p := i.Plan()
		product, _ := product.Get(p.Product.ID, nil)
		p.Product = product
		jsonPlan, _ := json.Marshal(p)
		_ = json.Unmarshal(jsonPlan, &plan)
		plans = append(plans, plan)
	}
	json.NewEncoder(w).Encode(plans)

}

func createCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var customError models.ErrorString

	token := r.Header.Get("token")

	if token == "" {
		customError.S = "Token cannot be empty"
		helper.GetError(&customError, w)
		return
	}

	userId, user_email, user_firstName, user_lastName, errorMsg := middleware.ValdateIncomingToken(token)
	if errorMsg != "" {
		customError.S = errorMsg
		helper.GetError(&customError, w)
		return
	}
	params := &stripe.CustomerParams{
		Description: stripe.String(user_firstName + user_lastName),
		Email:       stripe.String(user_email),
		Name:        stripe.String(user_firstName + user_lastName),
	}

	customer, err := customer.New(params)
	if err != nil {
		helper.GetError(err, w)
		return
	}
	fmt.Println("Customer ID: " + customer.ID)

	result := userCollection.FindOneAndUpdate(context.TODO(), bson.M{"user_id": userId}, bson.M{"$set": bson.M{"stripe_customerId": customer.ID}})
	// Decode result and write it to 'decoded'
	var decoded models.User
	err = result.Decode(&decoded)
	if err != nil {
		customError.S = "Cannot update User in database"
		helper.GetError(&customError, w)
		return
	}
	json.NewEncoder(w).Encode(customer)
}

var userSubscriptionCollection *mongo.Collection
var userCollection *mongo.Collection

func init() {
	//os.Setenv("STRIPE_SECRET_KEY", "sk_test_51JOIw2SDRD0D8UlDIme5WMEW2DvovQ61J11S9ppwKum0yeclDtdR3Uo5C3rI7Z6xgZ6R8XXOHD3ctclkEqkhUDMT001y0NN0da")
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	userSubscriptionCollection, userCollection = helper.ConnectDB()
}

func main() {
	r := mux.NewRouter()

	fmt.Println("Server started on port 3000")

	r.HandleFunc("/UserService/api/v1/subscriptions/plans", getPlans).Methods("GET")
	r.HandleFunc("/UserService/api/v1/subscriptions/customer", createCustomer).Methods("POST")

	// set our port address
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
