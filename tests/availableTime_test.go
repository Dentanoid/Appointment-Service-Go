package tests

import (
	"Group20/appointment-service/controllers"
	"Group20/appointment-service/database"
	"Group20/appointment-service/mqtt"
	"Group20/appointment-service/schemas"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var client MQTT.Client

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Error loading .env file")
	}
	database.Connect()
	client = mqtt.GetInstance()
	code := m.Run()
	os.Exit(code)
}

func TestAvailableTime(t *testing.T) {
	timeSlot := schemas.AvailableTime{
		ID:         primitive.NewObjectID(),
		Dentist_id: primitive.NewObjectID(),
		Start_time: primitive.DateTime(1704110400000),
		End_time:   primitive.DateTime(1704114000000),
		Clinic_id:  primitive.NewObjectID(),
	}
	var res controllers.Res

	result := controllers.CreateAvailableTime(timeSlot, res, client, false)
	if !result {
		t.Error("Error creating available time")
	}

	searchByDentist := schemas.AvailableTime{
		Dentist_id: timeSlot.Dentist_id,
	}
	result = controllers.GetAllAvailableTimes(searchByDentist, res, client)
	if !result {
		t.Error("Error finding available time by dentist")
	}

	searchByClinic := schemas.AvailableTime{
		Clinic_id: timeSlot.Clinic_id,
	}
	result = controllers.GetAllAvailableTimes(searchByClinic, res, client)
	if !result {
		t.Error("Error finding available time by clinic")
	}

	dentistArray := controllers.DentistArray{
		Clinics:    []primitive.ObjectID{timeSlot.Clinic_id},
		Start_time: timeSlot.Start_time,
		End_time:   timeSlot.End_time,
	}
	result = controllers.GetClinicsAvailabletimes(dentistArray, res, client)
	if !result {
		t.Error("Error finding available time by multiple clinics and time window")
	}

	result = controllers.DeleteAvailableTime(timeSlot.ID, res, client)
	if !result {
		t.Error("Error deleting available time")
	}

}
