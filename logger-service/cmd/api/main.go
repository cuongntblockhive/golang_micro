package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

const (
	webPort  = "80"
	rpcConn  = "5001"
	mongoUrl = "mongodb://mongo:27017"
	grpcPort = "50001"
)

var client *mongo.Client

func main() {
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panicln(err)
	}
	client = mongoClient
}

func connectToMongo() (*mongo.Client, error) {
	optionClient := options.Client().ApplyURI(mongoUrl)
	optionClient.SetAuth(options.Credential{
		Username: os.Getenv("MONGO_USER_NAME"),
		Password: os.Getenv("MONGO_PASSWORD"),
	})
	client, err := mongo.Connect(context.TODO(), optionClient)
	if err != nil {
		log.Panicln(err)
		return nil, err
	}
	return client, nil
}
