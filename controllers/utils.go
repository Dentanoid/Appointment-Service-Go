package controllers

import (
	"Group20/appointment-service/schemas"
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Res struct {
	Status         int                      `json:"status,omitempty"`
	RequestID      string                   `json:"requestID,omitempty"`
	Message        string                   `json:"message,omitempty"`
	AvailableTime  *schemas.AvailableTime   `json:"availabletime,omitempty"`
	Appointment    *schemas.Appointment     `json:"appointment,omitempty"`
	AvailableTimes *[]schemas.AvailableTime `json:"availabletimes,omitempty"`
	Appointments   *[]schemas.Appointment   `json:"appointments,omitempty"`
}

type DentistArray struct {
	RequestID  string               `json:"requestID,omitempty"`
	Clinics    []primitive.ObjectID `json:"clinics,omitempty"`
	Start_time primitive.DateTime   `json:"start_time,omitempty"`
	End_time   primitive.DateTime   `json:"end_time,omitempty"`
}

// Adds mqtt code to stringified json
func AddCodeStringJson(json string, code string) string {
	var newJson string
	length := len(json)
	index := 0

	runes := []rune(json)

	for index >= 0 && index < (length-1) {
		newJson = newJson + string(runes[index])
		index++
	}
	newJson = newJson + ",\"Code\": \"" + code + "\"}"
	return newJson
}

func PublishReturnMessage(returnData Res, topic string, client mqtt.Client) {

	returnJson, err := json.Marshal(returnData)
	if err != nil {
		panic(err)
	}

	returnString := string(returnJson)
	fmt.Printf(returnString)

	client.Publish(topic, 0, false, returnString)
}
