package controllers

import (
	"Group20/appointment-service/database"
	"Group20/appointment-service/schemas"
	"context"
	"encoding/json"
	"fmt"

	//"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitialiseAppointment(client mqtt.Client) {

	//CREATE
	tokenCancel := client.Subscribe("grp20/req/appointment/delete", byte(0), func(c mqtt.Client, m mqtt.Message) {

		var payload schemas.Appointment
		var returnData Res
		err1 := json.Unmarshal(m.Payload(), &payload)
		err2 := json.Unmarshal(m.Payload(), &returnData)
		if (err1 != nil) && (err2 != nil) {
			//send mqtt message 400 bad request
			fmt.Printf("400 - Bad request")
		} else {
			go CancelAppointment(payload.ID, returnData, client)
		}
	})
	if tokenCancel.Error() != nil {
		panic(tokenCancel.Error())
	}

	tokenCreate := client.Subscribe("grp20/req/appointment/create", byte(0), func(c mqtt.Client, m mqtt.Message) {

		var payload schemas.Appointment
		var returnData Res
		err1 := json.Unmarshal(m.Payload(), &payload)
		err2 := json.Unmarshal(m.Payload(), &returnData)
		if (err1 != nil) && (err2 != nil) {
			//send mqtt message 400 bad request
			fmt.Printf("400 - Bad request")
		} else {
			go CreateAppointment(payload, returnData, client)
		}
	})
	if tokenCreate.Error() != nil {
		panic(tokenCreate.Error())
	}

	tokenGetAllForUser := client.Subscribe("grp20/req/timeslots/get", byte(0), func(c mqtt.Client, m mqtt.Message) {

		var payload schemas.Appointment
		var returnData Res
		err1 := json.Unmarshal(m.Payload(), &payload)
		err2 := json.Unmarshal(m.Payload(), &returnData)
		if (err1 != nil) && (err2 != nil) {
			//send mqtt message 400 bad request
			fmt.Printf("400 - Bad request")
		} else {
			go GetAllForUser(payload, returnData, client)
		}
	})
	if tokenGetAllForUser.Error() != nil {
		panic(tokenGetAllForUser.Error())
	}
}

func DeleteAppointment(id primitive.ObjectID, returnData Res, client mqtt.Client) bool {
	var returnVal bool
	col := getAppointmentCollection()
	filter := bson.M{"_id": id}
	result, err := col.DeleteOne(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	if result.DeletedCount == 1 {

		returnData.Message = "Appointment deleted"
		returnData.Status = 200

	} else {
		returnData.Message = "Appointment not found"
		returnData.Status = 404
	}

	PublishReturnMessage(returnData, "grp20/res/timeslots/delete", client)

	returnVal = true
	return returnVal
}

func GetAllForUser(payload schemas.Appointment, returnData Res, client mqtt.Client) bool {

	var filter bson.M
	timeslots := make([]schemas.AvailableTime, 0)
	appointments := make([]schemas.Appointment, 0)

	var zeroID primitive.ObjectID
	if payload.Dentist_id != zeroID {
		filter = bson.M{"dentist_id": payload.Dentist_id}
	} else if payload.Patient_id != zeroID {
		filter = bson.M{"patient_id": payload.Patient_id}
	} else {
		returnData.Message = "Bad request"
		returnData.Status = 400
		PublishReturnMessage(returnData, "grp20/res/timeslots/get", client)
		return false
	}

	col := getAppointmentCollection()

	cursor, err := col.Find(context.TODO(), filter)

	if err != nil {
		returnData.Message = "Error"
		returnData.Status = 500
		PublishReturnMessage(returnData, "grp20/res/timeslots/get", client)
		return false
	}

	cursor.All(context.TODO(), &appointments)
	// defer cursor.Close(context.TODO())

	// for cursor.Next(context.TODO()) {
	// 	var appointment schemas.Appointment

	// 	if err := cursor.Decode(&appointment); err != nil {
	// 		returnData.Message = "Error"
	// 		returnData.Status = 500
	// 		PublishReturnMessage(returnData, "grp20/res/timeslots/get", client)
	// 		return false
	// 	}

	// 	timeslots = append(timeslots, appointment)
	// }

	col = getAvailableTimesCollection()

	cursor, err = col.Find(context.TODO(), filter)

	if err != nil {
		returnData.Message = "Error"
		returnData.Status = 500
		PublishReturnMessage(returnData, "grp20/res/timeslots/get", client)
		return false
	}

	cursor.All(context.TODO(), &timeslots)
	// defer cursor.Close(context.TODO())

	// for cursor.Next(context.TODO()) {
	// 	var availableTimes schemas.Appointment

	// 	if err := cursor.Decode(&availableTimes); err != nil {
	// 		returnData.Message = "Error"
	// 		returnData.Status = 500
	// 		PublishReturnMessage(returnData, "grp20/res/timeslots/get", client)
	// 		return false
	// 	}

	// 	timeslots = append(timeslots, availableTimes)
	// }

	if timeslots == nil {
		returnData.Status = 404
		returnData.Message = "No timeslots found"
	} else {
		returnData.Status = 200
		returnData.Appointments = &appointments
		returnData.AvailableTimes = &timeslots
	}

	PublishReturnMessage(returnData, "grp20/res/timeslots/get", client)

	return true
}

func CreateAppointment(payload schemas.Appointment, returnData Res, client mqtt.Client) bool {

	if appointmentExist(payload) {
		//send conflict - http status code
		return false
	}

	if payload.Start_time > payload.End_time {
		//send mqtt message, not valid format for start and end, should be start<end
		return false
	}

	col := getAppointmentCollection()

	result, err := col.InsertOne(context.TODO(), payload)
	payload.ID = result.InsertedID.(primitive.ObjectID)
	_ = result

	if err != nil {
		returnData.Status = 500
		returnData.Message = "Appointment could not be created"
		PublishReturnMessage(returnData, "grp20/res/availabletimes/book", client)
		return false
	}

	returnData.Message = "Appointment booked"
	returnData.Status = 200
	returnData.Appointment = &payload

	PublishReturnMessage(returnData, "grp20/res/availabletimes/book", client)
	return true
}

func CancelAppointment(id primitive.ObjectID, returnData Res, client mqtt.Client) bool {
	appointment := &schemas.Appointment{}

	var zeroID primitive.ObjectID

	col := getAppointmentCollection()
	filter := bson.M{"_id": id}
	data := col.FindOne(context.TODO(), filter)

	if data.Err() == mongo.ErrNoDocuments {
		//send 404 message
		returnData.Message = "Appointment not found"
		returnData.Status = 404
		PublishReturnMessage(returnData, "grp20/res/appointment/delete", client)

		return false
	}

	result, err := col.DeleteOne(context.TODO(), filter)

	if err != nil {
		panic(err)
	}

	if result.DeletedCount == 1 {

		returnData.Status = 200

		data.Decode(appointment)

		availableTime := schemas.AvailableTime{
			ID:         appointment.ID,
			Dentist_id: appointment.Dentist_id,
			Start_time: appointment.Start_time,
			End_time:   appointment.End_time,
			Clinic_id:  appointment.Clinic_id,
		}
		CreateAvailableTime(availableTime, returnData, client, true)

		appointment.ID = zeroID
		returnData.Appointment = appointment
		PublishReturnMessage(returnData, "grp20/req/booking/cancellation", client)
		PublishReturnMessage(returnData, "grp20/req/booking/cancellation/"+string(availableTime.Clinic_id.Hex()), client)

		return true
	} else {

		returnData.Status = 404
		returnData.Message = "Appointment not found"

		PublishReturnMessage(returnData, "grp20/res/appointment/delete", client)
		return false
	}

}

func appointmentExist(payload schemas.Appointment) bool {
	col := getAppointmentCollection()

	filter := bson.M{
		"dentist_id": payload.Dentist_id,
		"patient_id": payload.Patient_id,
		"start_time": payload.Start_time,
		"end_time":   payload.End_time,
	}

	count, err := col.CountDocuments(context.Background(), filter)
	if err != nil {
		return true
	}

	return count > 0
}

func getAppointmentCollection() *mongo.Collection {
	col := database.Database.Collection("Appointments")
	return col
}
