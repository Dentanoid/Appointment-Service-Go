package main

import (
	"Group20/appointment-service/database"
	"Group20/appointment-service/mqtt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {


	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to MongoDB
	database.Connect()

	// Connect to MQTT
    mqtt.GetInstance()

    defer database.Close()
    defer mqtt.Close()

	<-c
	// Variant #1 MQTT
	/*
		In the likely scenario that REST will only be adhered to between the client and the main gateway; The communication between this module will
		most likely happen through a more efficient protocol like MQTT or HTTP 2.0. In that case, an MQTT client can be found below
		which can subscribe and publish to different topics and handle events
	*/
	// fmt.Println("HTTP server is listening for requests...")
	// log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))

}
