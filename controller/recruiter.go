package controller

import (
	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/model"
	"github.com/FrosTiK-SD/auth/util"
	db "github.com/FrosTiK-SD/mongik/db"
	models "github.com/FrosTiK-SD/mongik/models"
	"go.mongodb.org/mongo-driver/bson"
)

func GetRecruiterByEmail(mongikClient *models.Mongik, email *string, role *string, noCache bool) (*model.RecruiterModelPopulated, *string) {
	var recruiterPopulated model.RecruiterModelPopulated
	recruiterPopulated, _ = db.AggregateOne[model.RecruiterModelPopulated](mongikClient, constants.DB, constants.COLLECTION_RECRUITER, []bson.M{{
		"$match": bson.M{
			"email": bson.M{
				"$regex":   "^" + *email + "$",
				"$options": "i",
			},
		},
	}, {
		"$lookup": bson.M{
			"from":         constants.COLLECTION_GROUP,
			"localField":   "groups",
			"foreignField": "_id",
			"as":           "groups",
		},
	}}, noCache)

	if !util.CheckRoleExists(&recruiterPopulated.GroupDetails, *role) {
		return nil, &constants.ERROR_NOT_A_RECRUITER
	}

	return &recruiterPopulated, nil
}
