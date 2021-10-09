FROM golang:1.16-alpine3.13 AS build_base
RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /tmp/user-subscription-service
COPY . .

RUN go mod download

# Build the Go app
RUN go build -o ./out/user-subscription-service .

# Start fresh from a smaller image
FROM alpine:3.9 
RUN apk add ca-certificates

COPY --from=build_base /tmp/user-subscription-service/out/user-subscription-service /app/user-subscription-service

# Set Env Variable
ENV MONGODB_URL=mongodb://mongoadmin:mongopassword@mongodb-0.database:27017/?authSource=admin
ENV STRIPE_SECRET_KEY=sk_test_51JQnzHSCDB2sPH6iDRL16h6sTgKFdz2VLcpBuhrCulCff2wdobRkdL0eTtuS43uic35RzsxkPi0szZqGNOIyBHBD0037uG9JMI

# Run the binary program produced by `go install`
CMD ["/app/user-subscription-service"]