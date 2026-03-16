package handler

import (
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	db "github.com/FrosTiK-SD/mongik/db"
	"github.com/FrosTiK-SD/auth/model"
	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/controller"
	"github.com/FrosTiK-SD/auth/util"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAllCompanies(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)

	companies, err := controller.GetAllCompanies(h.MongikClient, noCache)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_MONGO_ERROR,
			"message": err,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": companies,
	})
}

// Struct used for receiving recruiter and company data in request
type CreateRecruiterAndCompanyRequest struct {
	Company   model.Company          `json:"company" binding:"required"`
	Recruiter map[string]interface{} `json:"recruiter" binding:"required"`
}

var InitialGroupIDs = []string{"645afd0cfec4439851def4de"}

func batchConvertToObjectID(sArray []string) ([]primitive.ObjectID, error) {
	objIdArray := make([]primitive.ObjectID, 0, len(sArray))
	for _, s := range sArray {
		id, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			return nil, err
		}
		objIdArray = append(objIdArray, id)
	}
	return objIdArray, nil
}

func (h *Handler) CreateRecruiterAndCompany(ctx *gin.Context) {
	var req CreateRecruiterAndCompanyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()

	// Add timestamps to company struct (you can keep using the struct)
	req.Company.ID = primitive.NewObjectID() // optional, ensures clean _id
	companyDoc := req.Company
	companyMap := map[string]interface{}{
		"_id":                  companyDoc.ID,
		"name":                 companyDoc.Name,
		"logo":                 companyDoc.LogoURLs,
		"website":              companyDoc.Website,
		"address":              companyDoc.Address,
		"category":             companyDoc.Category,
		"sector":               companyDoc.Sector,
		"companyTurnover":      companyDoc.CompanyTurnover,
		"yearOfEstablishment":  companyDoc.YearOfEstablishment,
		"numberOfEmployees":    companyDoc.NumberOfEmployees,
		"createdAt":            now,
		"updatedAt":            now,
	}

	// Insert company into DB
	companyResult, err := db.InsertOne(h.MongikClient, constants.DB, constants.COLLECTION_COMPANY, companyMap)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	companyID := companyResult.InsertedID

	// Add required fields to recruiter object
	recruiterObj := req.Recruiter
	recruiterObj["isActive"] = true
	recruiterObj["company"] = companyID
	recruiterObj["createdAt"] = now
	recruiterObj["updatedAt"] = now

	// Add default group
	groups, err := batchConvertToObjectID(InitialGroupIDs)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Invalid group ID"})
		return
	}
	recruiterObj["groups"] = groups

	// Insert recruiter
	recruiterResult, err := db.InsertOne(h.MongikClient, constants.DB, constants.COLLECTION_RECRUITER, recruiterObj)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{
		"data": gin.H{
			"company":   companyResult,
			"recruiter": recruiterResult,
		},
	})
}
