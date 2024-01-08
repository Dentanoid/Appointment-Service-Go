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
		var dentistArray DentistArray

		var dateTime primitive.DateTime

		err1 := json.Unmarshal(m.Payload(), &payload)
		err2 := json.Unmarshal(m.Payload(), &returnData)
		err3 := json.Unmarshal(m.Payload(), &dentistArray)

		if ((err1 != nil) && (err3 != nil)) || (err2 != nil) {
			returnData.Message = "Bad request"
			returnData.Status = 400
			PublishReturnMessage(returnData, "grp20/res/availabletimes/get", client)
		} else if dentistArray.Start_time == dateTime {
			go GetAllAvailableTimes(payload, returnData, client)
		} else {
			go GetClinicsAvailabletimes(dentistArray, returnData, client)
		}
	})

	if tokenGet.Error() != nil {
		panic(tokenCreate.Error())
	}

	tokenDelete := client.Subscribe("grp20/req/timeslots/delete", byte(0), func(c mqtt.Client, m mqtt.Message) {
		var payload schemas.AvailableTime
		var returnData Res

		err1 := json.Unmarshal(m.Payload(), &payload)
		err2 := json.Unmarshal(m.Payload(), &returnData)
		if (err1 != nil) && (err2 != nil) {
			returnData.Message = "Bad request"
			returnData.Status = 400
			PublishReturnMessage(returnData, "grp20/res/timeslots/delete", client)
		} else {
			go DeleteAvailableTime(payload.ID, returnData, client)
		}
	})

	if tokenDelete.Error() != nil {
		panic(tokenCreate.Error())
	}

	tokenBookAvailableTime := client.Subscribe("grp20/req/availabletimes/book", byte(0), func(c mqtt.Client, m mqtt.Message) {
		var payload schemas.Appointment
		var returnData Res

		err1 := json.Unmarshal(m.Payload(), &payload)
		err2 := json.Unmarshal(m.Payload(), &returnData)
		if (err1 != nil) && (err2 != nil) {
			fmt.Printf("malformed payload!")
		} else {
			go BookAvailableTime(payload, returnData, client)
		}
	})

	if tokenBookAvailableTime.Error() != nil {
		panic(tokenBookAvailableTime.Error())
	}

}

func BookAvailableTime(payload schemas.Appointment, returnData Res, client mqtt.Client) bool {
	var deletedTime schemas.AvailableTime

	col := getAvailableTimesCollection()
	filter := bson.M{"_id": payload.ID}

	err := col.FindOneAndDelete(context.TODO(), filter).Decode(&deletedTime)
	if err != nil {
		returnData.Status = 500
		returnData.Message = "Internal server error"
		PublishReturnMessage(returnData, "grp20/res/availabletimes/book", client)
		return false
	}

	//if no document is found, deletedTime.ID will have an null (zero valued) _id
	if len(deletedTime.ID) == 0 {
		returnData.Status = 404
		returnData.Message = "Time slot not found"
		PublishReturnMessage(returnData, "grp20/res/availabletimes/book", client)
		return false
	}

	deletedTimeJson, err1 := json.Marshal(deletedTime)

	err2 := json.Unmarshal(deletedTimeJson, &payload)

	if (err1 != nil) || (err2 != nil) {
		returnData.Status = 500
		returnData.Message = "Internal server error"
		PublishReturnMessage(returnData, "grp20/res/availabletimes/book", client)
	}

	// var zeroObjectID primitive.ObjectID
	// payload.ID = zeroObjectID

	// Reinsert if creation of appointment is unsuccessfull
	if !CreateAppointment(payload, returnData, client) {
		result, err := col.InsertOne(context.TODO(), deletedTime)
		_ = result
		return err == nil
	} else {
		//If successfull, return an notification topic
		fmt.Println(payload)
		returnData.Appointment = &payload

		PublishReturnMessage(returnData, "grp20/req/booking/confirmation", client)
		PublishReturnMessage(returnData, "grp20/req/booking/confirmation/"+string(payload.Clinic_id.Hex()), client)
		return true
	}
}

func CreateAvailableTime(payload schemas.AvailableTime, returnData Res, client mqtt.Client, internal bool) bool {

	if exist(payload) {
		returnData.Message = "An identical available time already exist!"
		returnData.Status = 409
		if internal {
			PublishReturnMessage(returnData, "grp20/res/appointment/delete", client)
		} else {
			PublishReturnMessage(returnData, "grp20/res/availabletimes/create", client)
		}
		return false
	}

	if payload.Start_time > payload.End_time {
		returnData.Message = "End time must be after the start time"
		returnData.Status = 409
		if internal {
			PublishReturnMessage(returnData, "grp20/res/appointment/delete", client)
		} else {
			PublishReturnMessage(returnData, "grp20/res/availabletimes/create", client)
		}
		return false
	}

	col := getAvailableTimesCollection()


	result, err := col.InsertOne(context.TODO(), payload)

    insertedId, ok := result.InsertedID.(primitive.ObjectID)
    if !ok{
        fmt.Println("Could not get ID")
    }

    payload.ID = insertedId

    returnData.AvailableTime = &payload

	if internal == false {
		if err != nil {

			returnData.Message = "An error occurred"
			returnData.Status = 500
			PublishReturnMessage(returnData, "grp20/res/availabletimes/create", client)
			log.Fatal(err)
			return false

		}

		fmt.Printf("Registered availableTime with dentistID: %v \n", result.InsertedID)

		// Returns the time slot ID
		returnData.Message = "A new available time has been created"
		returnData.Status = 201
		PublishReturnMessage(returnData, "grp20/res/availabletimes/create", client)
		PublishReturnMessage(returnData, "grp20/availabletimes/live/"+string(payload.Clinic_id.Hex()), client)

		return true
	} else {
		if err != nil {
			//Data not migrated successfully
			return false
		} else {
			//Data migrated successfully, will get triggered when a patient cancels an appointment in appointment.go
			returnData.Message = "An appointment has been canceled"
			returnData.Status = 201
			PublishReturnMessage(returnData, "grp20/res/appointment/delete", client)
			PublishReturnMessage(returnData, "grp20/availabletimes/live/"+string(payload.Clinic_id.Hex()), client)
			return true
		}
	}
}

// GetAllAvailableTimes getAllInstancesWithDentistID retrieves all documents in a collection with a matching dentist_id
func GetAllAvailableTimes(payload schemas.AvailableTime, returnData Res, client mqtt.Client) bool {

	var filter bson.D
	col := getAvailableTimesCollection()

	var zeroID primitive.ObjectID
	if payload.Dentist_id != zeroID {
		filter = bson.D{{Key: "dentist_id", Value: payload.Dentist_id}}
	} else if payload.Clinic_id != zeroID {
		filter = bson.D{{Key: "clinic_id", Value: payload.Clinic_id}}
	} else {
		filter = bson.D{{}}
	}

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
	returnData.Status = 200

	returnData.AvailableTimes = &availableTimes

	PublishReturnMessage(returnData, "grp20/res/availabletimes/get", client)

	return true
}

// getting all availabletimes based on an array of clinic_id and a time window consisting of start_time and end_time
func GetClinicsAvailabletimes(payload DentistArray, returnData Res, client mqtt.Client) bool {
	col := getAvailableTimesCollection()

	var availableTimes []schemas.AvailableTime

	for _, clinicID := range payload.Clinics {
		// Define the filter based on the provided criteria
		filter := bson.M{
			"clinic_id":  clinicID,
			"start_time": bson.M{"$gte": payload.Start_time},
			"end_time":   bson.M{"$lte": payload.End_time},
		}

		cursor, err := col.Find(context.TODO(), filter)
		if err != nil {
			returnData.Message = "An error occurred"
			returnData.Status = 500
			PublishReturnMessage(returnData, "grp20/res/availabletimes/get", client)
			return false
		}

		defer cursor.Close(context.TODO())

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
	}

	// Convert the responseMap to JSON
	returnData.AvailableTimes = &availableTimes
	returnData.Message = "Available times successfully fetched!"
	returnData.Status = 200
	PublishReturnMessage(returnData, "grp20/res/availabletimes/get", client)

	return true
}

// deletes an availableTime entirely, will be performed by dentists
func DeleteAvailableTime(ID primitive.ObjectID, returnData Res, client mqtt.Client) bool {

	col := getAvailableTimesCollection()
	fmt.Print(ID.Hex())
	filter := bson.M{"_id": ID}
	result, err := col.DeleteOne(context.TODO(), filter)

	if err != nil {
		log.Fatal(err)
		return false
	}

	if result.DeletedCount != 1 {

		if err != nil {

			returnData.Message = "Internal server error!"
			returnData.Status = 500
			PublishReturnMessage(returnData, "grp20/res/timeslots/delete", client)

			return false
		}
		//if no availabletime is found for _id, check in appointments collection
		DeleteAppointment(ID, returnData, client)

		return false
	} else {
		fmt.Printf("Deleted Time id: %v \n", ID)

		returnData.Message = "Available time deleted!"
		returnData.Status = 200
		PublishReturnMessage(returnData, "grp20/res/timeslots/delete", client)

		return true

	}
}

func exist(payload schemas.AvailableTime) bool {
	col := getAvailableTimesCollection()

	filter := bson.M{
		"dentist_id": payload.Dentist_id,
		"start_time": payload.Start_time,
		"end_time":   payload.End_time,
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
