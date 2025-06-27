package handler

import (
	"net/http"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/controller"
	"github.com/FrosTiK-SD/auth/model"
	"github.com/FrosTiK-SD/auth/util"
	db "github.com/FrosTiK-SD/mongik/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

type CreateRecruiterAndCompanyRequest struct {
	Company   model.Company         `json:"company" binding:"required"`
	Recruiter map[string]interface{} `json:"recruiter" binding:"required"` // Adjust type as needed
}

var InitialGroupIDs = []string{"645afd0cfec4439851def4de"} // You can move this to constants if needed

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
	// 1. Insert company
	companyResult, err := db.InsertOne(h.MongikClient, constants.DB, constants.COLLECTION_COMPANY, req.Company)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	companyID := companyResult.InsertedID

	// 2. Prepare recruiter object
	recruiterObj := req.Recruiter
	recruiterObj["isActive"] = true
	recruiterObj["company"] = companyID

	groups, err := batchConvertToObjectID([]string{"645afd0cfec4439851def4de"})
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Invalid group ID"})
		return
	}
	recruiterObj["groups"] = groups

	// 3. Insert recruiter
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
