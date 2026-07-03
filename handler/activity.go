package handler
import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/model"
	db "github.com/FrosTiK-SD/mongik/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
type LogEntryPopulated struct {
	Id          primitive.ObjectID   `json:"_id" bson:"_id"`
	Type        string               `json:"type" bson:"type"`
	Timestamp   primitive.DateTime   `json:"timestamp" bson:"timestamp"`
	User        primitive.ObjectID   `json:"user" bson:"user"`
	Message     string               `json:"message" bson:"message"`
	Ref         *primitive.DBPointer `json:"ref" bson:"ref"`
	CreatedAt   primitive.DateTime   `json:"createdAt" bson:"createdAt"`
	UpdatedAt   primitive.DateTime   `json:"updatedAt" bson:"updatedAt"`
	UserDetails *struct {
		Id        primitive.ObjectID `json:"_id" bson:"_id"`
		Email     string             `json:"email" bson:"email"`
		FirstName string             `json:"firstName" bson:"firstName"`
		LastName  *string            `json:"lastName" bson:"lastName"`
	} `json:"user_details" bson:"user_details"`
}
type LogResponse struct {
	Metadata []struct {
		Total int `json:"total" bson:"total"`
	} `json:"metadata" bson:"metadata"`
	Data []LogEntryPopulated `json:"data" bson:"data"`
}
func (h *Handler) GetActivityLogs(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "50")
	skipStr := ctx.DefaultQuery("skip", "0")
	query := ctx.Query("query")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	skip, err := strconv.Atoi(skipStr)
	if err != nil || skip < 0 {
		skip = 0
	}
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         constants.COLLECTION_STUDENT,
				"localField":   "user",
				"foreignField": "_id",
				"as":           "user_details",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$user_details",
				"preserveNullAndEmptyArrays": true,
			},
		},
	}
	if strings.TrimSpace(query) != "" {
		regexQuery := primitive.Regex{Pattern: query, Options: "i"}
		matchFilter := bson.M{
			"$or": []bson.M{
				{"message": regexQuery},
				{"user_details.email": regexQuery},
				{"user_details.firstName": regexQuery},
				{"user_details.lastName": regexQuery},
			},
		}
		pipeline = append(pipeline, bson.M{"$match": matchFilter})
	}
	pipeline = append(pipeline, bson.M{"$sort": bson.M{"timestamp": -1}})
	facetPipeline := append(pipeline, bson.M{
		"$facet": bson.M{
			"metadata": []bson.M{
				{"$count": "total"},
			},
			"data": []bson.M{
				{"$skip": skip},
				{"$limit": limit},
			},
		},
	})
	results, err := db.Aggregate[LogResponse](h.MongikClient, constants.DB, "activities", facetPipeline, true)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	total := 0
	data := []LogEntryPopulated{}
	if len(results) > 0 {
		if len(results[0].Metadata) > 0 {
			total = results[0].Metadata[0].Total
		}
		if results[0].Data != nil {
			data = results[0].Data
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"total": total,
		"data":  data,
	})
}
func (h *Handler) CreateActivityLog(ctx *gin.Context) {
	value, exists := ctx.Get(constants.SESSION)
	student, ok := value.(*model.StudentPopulated)
	if !exists || !ok {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}
	var req struct {
		Type    string `json:"type" binding:"required"`
		Message string `json:"message" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.LogActivityDirect(student.Id, req.Type, req.Message)
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
func (h *Handler) LogActivityDirect(user primitive.ObjectID, activityType string, message string) {
	now := primitive.NewDateTimeFromTime(time.Now())
	activity := bson.M{
		"_id":       primitive.NewObjectID(),
		"type":      activityType,
		"timestamp": now,
		"user":      user,
		"message":   message,
		"createdAt": now,
		"updatedAt": now,
	}
	_, err := db.InsertOne(h.MongikClient, constants.DB, "activities", activity)
	if err != nil {
		fmt.Printf("Error inserting activity log: %v\n", err)
	}
}
