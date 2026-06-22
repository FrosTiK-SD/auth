package interfaces

import (
	Constant "github.com/FrosTiK-SD/models/constant"
	Student "github.com/FrosTiK-SD/models/student"
)

type StudentRegistration struct {
	FirstName  string  `json:"firstName" bson:"firstName"`
	MiddleName *string `json:"middleName" bson:"middleName"`
	LastName   *string `json:"lastName" bson:"lastName"`

	Batch          Student.Batch `json:"batch" bson:"batch"`
	RollNo         int           `json:"rollNo" bson:"rollNo"`
	InstituteEmail string        `json:"email" bson:"email"`
	Department     string        `json:"department" bson:"department"`
	Course         string        `json:"course" bson:"course"`
	Specialisation *string       `json:"specialisation" bson:"specialisation"`

	Mobile        string           `json:"mobile" bson:"mobile"`
	PersonalEmail string           `json:"personalEmail" bson:"personalEmail"`
	Gender        *Constant.Gender `json:"gender" bson:"gender"`

	RawKeyStore map[string]interface{} `json:"raw_key_store" bson:"raw_key_store"`
}

type UnverifyBatchRequest struct {
	StartYear int `json:"startYear" bson:"startYear" binding:"required"`
	EndYear   int `json:"endYear" bson:"endYear" binding:"required"`
}