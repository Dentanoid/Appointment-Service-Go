package schemas

import "go.mongodb.org/mongo-driver/bson/primitive"

type Appointment struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Patient_id primitive.ObjectID `bson:"patient_id" json:"patient_id"`
	Dentist_id primitive.ObjectID `bson:"dentist_id" json:"dentist_id"`
	Start_time primitive.DateTime `bson:"start_time" json:"start_time"`
	End_time   primitive.DateTime `bson:"end_time" json:"end_time"`
}
