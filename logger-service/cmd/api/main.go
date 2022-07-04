package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"logger/data"
	"net/http"
	"os"
	"time"
)

const (
	webPort = "80"
	//rpcConn  = "5001"
	mongoUrl = "mongodb://localhost:27017"
	//grpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panicln(err)
	}
	client = mongoClient

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}
	log.Println("Starting service on port", webPort)
	app.serve()
}

func (app *Config) serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	// start server
	err := srv.ListenAndServe()

	if err != nil {
		log.Panicf("Err run server")
	}
	log.Printf("ListenAndServe")
}

func connectToMongo() (*mongo.Client, error) {
	optionClient := options.Client().ApplyURI(mongoUrl)
	optionClient.SetAuth(options.Credential{
		Username: os.Getenv("MONGO_USER_NAME"),
		Password: os.Getenv("MONGO_PASSWORD"),
	})
	client, err := mongo.Connect(context.TODO(), optionClient)
	if err != nil {
		log.Println("Error connecting:", err)
		return nil, err
	}
	log.Println("Connected to mongo!!")
	return client, nil
}
