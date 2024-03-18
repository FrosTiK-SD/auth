package interfaces

import (
	Student "github.com/FrosTiK-SD/models/student"
)

type StudentRegistration struct {
	Batch          Student.Batch `json:"batch" bson:"batch"`
	RollNo         int           `json:"rollNo" bson:"rollNo"`
	InstituteEmail string        `json:"email" bson:"email"`
	Department     string        `json:"department" bson:"department"`
	Course         string        `json:"course" bson:"course"`
	Specialisation *string       `json:"specialisation" bson:"specialisation"`

	FirstName  string  `json:"firstName" bson:"firstName"`
	MiddleName *string `json:"middleName" bson:"middleName"`
	LastName   *string `json:"lastName" bson:"lastName"`
}
