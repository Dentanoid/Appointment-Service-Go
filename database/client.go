package database

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Database *mongo.Database

func Connect() {
	c, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		panic(err)
	}
	Database = c.Database("AppointmentService")
	fmt.Println("App is connected to MongoDB")
}

func Close() {
    if Database != nil{ 
        Database.Client().Disconnect(context.TODO()) 
        Database = nil
        fmt.Println("Database connection closed") 
    }
}
