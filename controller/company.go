package controller

import (
	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/model"
	db "github.com/FrosTiK-SD/mongik/db"
	mongikModels "github.com/FrosTiK-SD/mongik/models"
	"go.mongodb.org/mongo-driver/bson"
)

func GetAllCompanies(mongikClient *mongikModels.Mongik, noCache bool, currentPage int, companiesPerPage int) ([]model.Company, int, error) {

	var pipeline []bson.M

	pipeline = []bson.M{
		{"$sort": bson.M{"createdAt": -1}},
	}

	if currentPage != 0 {
		skip := (currentPage - 1) * companiesPerPage
		limit := companiesPerPage

		pipeline = []bson.M{
			{"$sort": bson.M{"createdAt": -1}},
			{"$skip": skip},
			{"$limit": limit},
		}
	}

	totalCompaniesArray, err := db.Aggregate[map[string]int](mongikClient, constants.DB, constants.COLLECTION_COMPANY, []bson.M{{"$count": "total"}}, noCache)
	if err != nil {
		return nil, 0, err
	}

	totalCompanies := totalCompaniesArray[0]["total"]

	companies, err := db.Aggregate[model.Company](mongikClient, constants.DB, constants.COLLECTION_COMPANY, pipeline, noCache)

	return companies, totalCompanies, err
}
