package handler

import (
	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/controller"
	"github.com/FrosTiK-SD/auth/interfaces"
	"github.com/FrosTiK-SD/models/constant"
	studentModel "github.com/FrosTiK-SD/models/student"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) HandlerUpdateStudentDetails(ctx *gin.Context) {
	studentCollection := h.MongikClient.MongoClient.Database(constants.DB).Collection(constants.COLLECTION_STUDENT)
	updatedStudent := studentModel.Student{}
	if errBinding := ctx.ShouldBindBodyWith(&updatedStudent, binding.JSON); errBinding != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": errBinding.Error()})
		return
	}

	filter := bson.M{"_id": h.Session.Student.Id, "email": h.Session.Student.InstituteEmail}
	currentStudent := studentModel.Student{}

	if errFind := studentCollection.FindOne(ctx, filter).Decode(&currentStudent); errFind != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": errFind.Error()})
		return
	}

	controller.AssignUnVerifiedFields(&updatedStudent, &currentStudent)
	controller.InvalidateVerifiedFieldsOnChange(&updatedStudent, &currentStudent)

	if updateResult, errUpdate := studentCollection.ReplaceOne(ctx, filter, currentStudent); errUpdate != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": errUpdate.Error()})
		return
	} else {
		ctx.JSON(200, gin.H{"student": updateResult})
	}
}

func (h *Handler) HandlerRegisterStudentDetails(ctx *gin.Context) {
	idToken := ctx.GetHeader("token")
	studentCollection := h.MongikClient.MongoClient.Database(constants.DB).Collection(constants.COLLECTION_STUDENT)
	newStudentDetails := interfaces.StudentRegistration{}

	if errBinding := ctx.ShouldBindBodyWith(&newStudentDetails, binding.JSON); errBinding != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": errBinding.Error()})
		return
	}

	if email, _, errVerify := controller.VerifyToken(h.MongikClient.CacheClient, idToken, h.JwkSet, true); errVerify != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": errVerify})
		return
	} else if email != &newStudentDetails.InstituteEmail {
		ctx.AbortWithStatusJSON(401, gin.H{"error": "Email mismatch"})
	}

	newStudent := studentModel.Student{
		Id:             primitive.NewObjectID(),
		Batch:          &newStudentDetails.Batch,
		RollNo:         newStudentDetails.RollNo,
		InstituteEmail: newStudentDetails.InstituteEmail,
		Department:     newStudentDetails.Department,
		Course:         (*constant.Course)(&newStudentDetails.Course),
		Specialisation: newStudentDetails.Specialisation,
		FirstName:      newStudentDetails.FirstName,
		MiddleName:     newStudentDetails.MiddleName,
		LastName:       newStudentDetails.LastName,
	}

	if result, err := studentCollection.InsertOne(ctx, newStudent); err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
		return
	} else {
		ctx.JSON(200, gin.H{"student": newStudent, "logs": result})
		return
	}
}
