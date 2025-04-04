package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/controller"
	"github.com/FrosTiK-SD/auth/interfaces"
	"github.com/FrosTiK-SD/auth/model"
	"github.com/FrosTiK-SD/auth/util"
	"github.com/FrosTiK-SD/models/constant"
	studentModel "github.com/FrosTiK-SD/models/student"
	db "github.com/FrosTiK-SD/mongik/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetStudentRoleObjectID() primitive.ObjectID {
	if objID, err := primitive.ObjectIDFromHex(os.Getenv(constants.ENV_STUDENT_GROUP_OBJ_ID)); err != nil {
		return primitive.NilObjectID
	} else {
		return objID
	}
}

func (h *Handler) GetAllStudents(ctx *gin.Context) {

	noCache := util.GetNoCache(ctx)

	students, err := controller.GetAllStudents(h.MongikClient, noCache)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"data":  nil,
			"error": err,
		})
	}
	ctx.JSON(http.StatusOK,
		gin.H{
			"data":  students,
			"error": nil,
		})
}

func (h *Handler) GetStudentById(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)
	_id, err := primitive.ObjectIDFromHex(ctx.GetHeader("id"))

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invaild ObjectID",
		})
		return
	}

	student, err := controller.GetStudentById(h.MongikClient, _id, noCache)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Could Not Fetch Student",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": student,
	})

}

func (h *Handler) GetAllTprs(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)
	tprs, err := controller.GetAllStudentsOfRole(h.MongikClient, constants.ROLE_TPR, noCache)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Could Not Fetch TPRs",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": tprs,
	})
}

// Already verified as Tpr by middleware
func (h *Handler) HandlerTprLogin(ctx *gin.Context) {
	value, exists := ctx.Get(constants.SESSION)
	student, ok := value.(*model.StudentPopulated)

	if !exists || !ok {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": constants.ERROR_ROLE_CHECK_FAILED,
			"error":   "Student does not exist",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": student,
	})
}

func (h *Handler) HandlerUpdateStudentDetails(ctx *gin.Context) {
	studentCollection := h.MongikClient.MongoClient.Database(constants.DB).Collection(constants.COLLECTION_STUDENT)

	student, exists := ctx.Get(constants.SESSION)
	if !exists {
		ctx.AbortWithStatusJSON(401, gin.H{"error": "Cant get student"})
		return
	}

	studentPopulated := student.(*model.StudentPopulated)

	var updatedStudent studentModel.Student
	if errBinding := ctx.ShouldBindJSON(&updatedStudent); errBinding != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": errBinding.Error()})
		return
	}

	filter := bson.M{"_id": studentPopulated.Id, "email": studentPopulated.InstituteEmail}

	var currentStudent studentModel.Student
	if errFind := studentCollection.FindOne(ctx, filter).Decode(&currentStudent); errFind != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": errFind.Error()})
		return
	}

	controller.AssignUnVerifiedFields(&updatedStudent, &currentStudent)
	controller.InvalidateVerifiedFieldsOnChange(&updatedStudent, &currentStudent)

	if updateResult, errUpdate := db.ReplaceOne(h.MongikClient, constants.DB, constants.COLLECTION_STUDENT, filter, &currentStudent); errUpdate != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": errUpdate.Error()})
		return
	} else {
		ctx.JSON(200, gin.H{"student": updateResult})
	}
}

func (h *Handler) HandlerRegisterStudentDetails(ctx *gin.Context) {
	idToken := ctx.GetHeader("token")
	newStudentDetails := interfaces.StudentRegistration{}

	if errBinding := ctx.BindJSON(&newStudentDetails); errBinding != nil {
		return
	}

	if email, _, errVerify := controller.VerifyToken(h.MongikClient.CacheClient, idToken, h.JwkSet, true); errVerify != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": errVerify})
		return
	} else {
		if !util.CheckValidInstituteEmail(*email) {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "not a valid institute email"})
			return
		}
		newStudentDetails.InstituteEmail = *email
	}

	newStudent := studentModel.Student{
		Groups:         []primitive.ObjectID{GetStudentRoleObjectID()},
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
		PersonalEmail:  newStudentDetails.PersonalEmail,
		Mobile:         newStudentDetails.Mobile,
		Gender:         newStudentDetails.Gender,
		CreatedAt:      primitive.NewDateTimeFromTime(time.Now().UTC()),
		UpdatedAt:      primitive.NewDateTimeFromTime(time.Now().UTC()),
	}

	if result, err := db.InsertOne(h.MongikClient, constants.DB, constants.COLLECTION_STUDENT, newStudent); err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
		return
	} else {
		ctx.JSON(200, gin.H{"student": newStudent, "logs": result})
		return
	}
}

func (h *Handler) HandlerGetStudentProfile(ctx *gin.Context) {
	student, exists := ctx.Get(constants.SESSION)
	if !exists {
		ctx.AbortWithStatusJSON(401, gin.H{"error": "Cannot get student"})
		return
	}

	studentPopulated := student.(*model.StudentPopulated)
	studentProfile := interfaces.StudentProfile{}
	controller.MapStudentToStudentProfile(&studentProfile, &studentPopulated.Student, true)

	ctx.JSON(200, gin.H{"profile": studentProfile, "isAcademicsVerified": studentPopulated.Academics.Verification.IsVerified})
}

func (h *Handler) HandlerUpdateStudentProfile(ctx *gin.Context) {
	_, exists := ctx.Get(constants.SESSION)
	if !exists {
		ctx.AbortWithStatusJSON(401, gin.H{"error": "Cannot get student"})
		return
	}

	updatedStudent := studentModel.Student{}
	studentProfile := interfaces.StudentProfile{}

	ctx.BindJSON(&studentProfile)
	controller.MapStudentToStudentProfile(&studentProfile, &updatedStudent, false)

	filter := bson.M{"email": updatedStudent.InstituteEmail}

	var currentStudent studentModel.Student
	studentCollection := h.MongikClient.MongoClient.Database(constants.DB).Collection(constants.COLLECTION_STUDENT)
	if errFind := studentCollection.FindOne(ctx, filter).Decode(&currentStudent); errFind != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": errFind.Error()})
		return
	}

	controller.AssignUnVerifiedFields(&updatedStudent, &currentStudent)
	controller.InvalidateVerifiedFieldsOnChange(&updatedStudent, &currentStudent)

	if updateResult, errUpdate := db.ReplaceOne(h.MongikClient, constants.DB, constants.COLLECTION_STUDENT, filter, &currentStudent); errUpdate != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": errUpdate.Error()})
		return
	} else {
		ctx.JSON(200, gin.H{"student": updateResult})
	}

	ctx.JSON(200, gin.H{"student": currentStudent})
}
