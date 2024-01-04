package tests

import (
	"Group20/appointment-service/controllers"
	"Group20/appointment-service/schemas"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAppointment(t *testing.T) {
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
		t.Error("Error creating available time for booking")
	}

	appointment := schemas.Appointment{
		ID:         timeSlot.ID,
		Patient_id: primitive.NewObjectID(),
	}
	result = controllers.BookAvailableTime(appointment, res, client)
	if !result {
		t.Error("Error booking available time")
	}

	searchByDentist := schemas.Appointment{
		Dentist_id: timeSlot.ID,
	}
	result = controllers.GetAllForUser(searchByDentist, res, client)
	if !result {
		t.Error("Error retrieving timeslots by dentist")
	}

	searchByPatient := schemas.Appointment{
		Patient_id: timeSlot.ID,
	}
	result = controllers.GetAllForUser(searchByPatient, res, client)
	if !result {
		t.Error("Error retrieving timeslots by patient id")
	}

	result = controllers.CancelAppointment(timeSlot.ID, res, client)
	if !result {
		t.Error("Error canceling appointment")
	}

	result = controllers.BookAvailableTime(appointment, res, client)
	if !result {
		t.Error("Error booking available time")
	}

	result = controllers.DeleteAppointment(timeSlot.ID, res, client)
	if !result {
		t.Error("Error deleting appointment")
	}

}
