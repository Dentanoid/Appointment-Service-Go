package schemas

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AvailableTime struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Dentist_id primitive.ObjectID `bson:"dentist_id" json:"dentist_id"`
	Start_time primitive.DateTime `bson:"start_time" json:"start_time"`
	End_time   primitive.DateTime `bson:"end_time" json:"end_time"`
	Clinic_id  primitive.ObjectID `bson:"clinic_id" json:"clinic_id"`
}
