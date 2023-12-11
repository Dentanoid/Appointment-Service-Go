package controllers

import (
	"Group20/appointment-service/database"
	"Group20/appointment-service/schemas"
	"context"
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeAvailableTimes(client mqtt.Client) {

	tokenCreate := client.Subscribe("grp20/req/availabletimes/create", byte(0), func(c mqtt.Client, m mqtt.Message) {
		var payload schemas.AvailableTime
		var returnData Res

		err1 := json.Unmarshal(m.Payload(), &payload)
		err2 := json.Unmarshal(m.Payload(), &returnData)
		if (err1 != nil) && (err2 != nil) {
			returnData.Message = "Bad Request"
			returnData.Status = 400
			PublishReturnMessage(returnData, "grp20/res/availabletimes/create", client)
		} else {
			go CreateAvailableTime(payload, returnData, client, false)
		}
	})

	if tokenCreate.Error() != nil {
		panic(tokenCreate.Error())
	}

	tokenGet := client.Subscribe("grp20/req/availabletimes/get", byte(0), func(c mqtt.Client, m mqtt.Message) {
		var payload schemas.AvailableTime
		var returnData Res

		err1 := json.Unmarshal(m.Payload(), &payload)
		err2 := json.Unmarshal(m.Payload(), &returnData)
		if (err1 != nil) && (err2 != nil) {
			returnData.Message = "Bad request"
			returnData.Status = 400
			PublishReturnMessage(returnData, "grp20/res/availabletimes/get", client)
		} else {
			go GetAllAvailableTimesWithDentistID(payload.Dentist_id, returnData, client)
		}
	})

	if tokenGet.Error() != nil {
		panic(tokenCreate.Error())
	}

	tokenDelete := client.Subscribe("grp20/req/availabletimes/delete", byte(0), func(c mqtt.Client, m mqtt.Message) {
		var payload schemas.AvailableTime
		var returnData Res

		err1 := json.Unmarshal(m.Payload(), &payload)
		err2 := json.Unmarshal(m.Payload(), &returnData)
		if (err1 != nil) && (err2 != nil) {
			returnData.Message = "Bad request"
			returnData.Status = 400
			PublishReturnMessage(returnData, "grp20/res/availabletimes/delete", client)
		} else {
			go DeleteAvailableTime(payload.ID, returnData, client)
		}
	})

	if tokenDelete.Error() != nil {
		panic(tokenCreate.Error())
	}

	tokenInternalMigrate := client.Subscribe("appointmentservice/internal/migrate", byte(0), func(c mqtt.Client, m mqtt.Message) {
		var payload schemas.AvailableTime
		var returnData Res

		err1 := json.Unmarshal(m.Payload(), &payload)
		err2 := json.Unmarshal(m.Payload(), &returnData)
		if (err1 != nil) && (err2 != nil) {
			fmt.Printf("malformed payload!")
		} else {
			go CreateAvailableTime(payload, returnData, client, true)
		}
	})

	if tokenInternalMigrate.Error() != nil {
		panic(tokenInternalMigrate.Error())
	}

    tokenBookAvailableTime := client.Subscribe("grp20/req/availabletimes/book", byte(0), func(c mqtt.Client, m mqtt.Message) {
	    var payload schemas.Appointment
        var returnData Res

	    err1 := json.Unmarshal(m.Payload(), &payload)
	    err2 := json.Unmarshal(m.Payload(), &returnData)
	    if (err1 != nil) && (err2 != nil){
	    	fmt.Printf("malformed payload!")
	    } else {
	    	go BookAvailableTime(payload, returnData, client)
        }
	})

	if tokenBookAvailableTime.Error() != nil {
		panic(tokenBookAvailableTime.Error())
	}


}

func BookAvailableTime(payload schemas.Appointment, returnData Res, client mqtt.Client) bool{
    var deletedTime schemas.AvailableTime

    col := getAvailableTimesCollection()
    filter := bson.M{"_id": payload.ID}  

    err := col.FindOneAndDelete(context.TODO(), filter).Decode(&deletedTime)
    if err != nil {
    //TODO send error message
        returnData.Status = 404
        returnData.Message = "Time slot not found"
        PublishReturnMessage(returnData, "grp20/res/availabletimes/book", client)
        return false
    }

    
    deletedTimeJson, err1 := json.Marshal(deletedTime)

    err2 := json.Unmarshal(deletedTimeJson, &payload)

    if (err1 != nil) || (err2 != nil) {
        //TODO send mqtt error
    }

    var zeroObjectID primitive.ObjectID
    payload.ID = zeroObjectID

    //Reinsert if creation of appointment is unsuccessfull
    if !CreateAppointment(payload, returnData, client) {
        result, err := col.InsertOne(context.TODO(), deletedTime) 
        _ = result
        return err == nil 
    }
    return true
}

func CreateAvailableTime(payload schemas.AvailableTime, returnData Res, client mqtt.Client, internal bool) bool {

	if exist(payload) {
		returnData.Message = "An identical available time already exist!"
		returnData.Status = 409
		PublishReturnMessage(returnData, "grp20/res/availabletimes/create", client)
		return false
	}

	if payload.Start_time > payload.End_time {
		returnData.Message = "End time must be after the start time"
		returnData.Status = 409
		PublishReturnMessage(returnData, "grp20/res/availabletimes/create", client)
		return false
	}

	col := getAvailableTimesCollection()

	result, err := col.InsertOne(context.TODO(), payload)

	if internal == false {
		if err != nil {
			log.Fatal(err)

			returnData.Message = "An error occurred"
			returnData.Status = 500
			PublishReturnMessage(returnData, "grp20/res/availabletimes/create", client)

			return false
		}

		fmt.Printf("Registered availableTime with dentistID: %v \n", result.InsertedID)

		// Returns the time slot ID
		returnData.Message = result.InsertedID.(primitive.ObjectID).Hex()
		returnData.Status = 201
		PublishReturnMessage(returnData, "grp20/res/availabletimes/create", client)

		return true
	} else {
		if err != nil {
			//Data not migrated successfully
			return false
		} else {
			//Data migrated successfully
			return true
		}
	}
}

// getAllInstancesWithDentistID retrieves all documents in a collection with a matching dentist_id
func GetAllAvailableTimesWithDentistID(dentistID primitive.ObjectID, returnData Res, client mqtt.Client) bool {

	col := getAvailableTimesCollection()
	filter := bson.D{{Key: "dentist_id", Value: dentistID}}

	cursor, err := col.Find(context.TODO(), filter)
	if err != nil {

		returnData.Message = "An error occurred"
		returnData.Status = 500
		PublishReturnMessage(returnData, "grp20/res/availabletimes/get", client)

		return false
	}

	defer cursor.Close(context.TODO())

	var availableTimes []schemas.AvailableTime

	for cursor.Next(context.TODO()) {
		var availableTime schemas.AvailableTime
		if err := cursor.Decode(&availableTime); err != nil {

			returnData.Message = "An error occurred while decoding results"
			returnData.Status = 500
			PublishReturnMessage(returnData, "grp20/res/availabletimes/get", client)

			return false
		}
		availableTimes = append(availableTimes, availableTime)
	}

	if err := cursor.Err(); err != nil {

		returnData.Message = "An error occurred"
		returnData.Status = 500
		PublishReturnMessage(returnData, "grp20/res/availabletimes/get", client)

		return false
	}

	// Convert the responseMap to JSON
	returnData.AvailableTimes = &availableTimes

	PublishReturnMessage(returnData, "grp20/res/availabletimes/get", client)

	return true
}

// deletes an availableTime entirely, will be performed by dentists
func DeleteAvailableTime(ID primitive.ObjectID, returnData Res, client mqtt.Client) bool {

	col := getAvailableTimesCollection()
	filter := bson.M{"_id": ID}
	result, err := col.DeleteOne(context.TODO(), filter)
	_ = result

	if err != nil {
		log.Fatal(err)
		return false
	}

	msg := schemas.AvailableTime{
		ID: ID,
	}

	if result.DeletedCount == 0 {
		document, err := json.Marshal(msg)

		if err != nil {

			returnData.Message = "Internal server error!"
			returnData.Status = 500
			PublishReturnMessage(returnData, "grp20/res/availabletimes/delete", client)

			return false
		}

		client.Publish("appointmentservice/internal/delete", 0, false, document)

		return false
	} else {
		fmt.Printf("Deleted Time id: %v \n", ID)

		returnData.Message = "Available time deleted!"
		returnData.Status = 200
		PublishReturnMessage(returnData, "grp20/res/availabletimes/delete", client)

		return true

	}
}

func exist(payload schemas.AvailableTime) bool {
	col := getAvailableTimesCollection()

	filter := bson.M{
		"Dentist_id": payload.Dentist_id,
		"Start_time": payload.Start_time,
		"End_time":   payload.End_time,
	}

	count, err := col.CountDocuments(context.Background(), filter)
	if err != nil {
		return false
	}

	return count > 0
}

func getAvailableTimesCollection() *mongo.Collection {
	col := database.Database.Collection("AvailableTimes")
	return col
}
