package handler

import (
	"bytes"
	"encoding/csv"
	stdjson "encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"fmt"

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

func csvBoolText(value bool) string {
	if value {
		return "TRUE"
	}
	return "FALSE"
}

func csvStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func csvDateTime(value primitive.DateTime) string {
	if value == 0 {
		return ""
	}
	return value.Time().UTC().Format(time.RFC3339)
}

func csvDateTimePtr(value *primitive.DateTime) string {
	if value == nil {
		return ""
	}
	return csvDateTime(*value)
}

func csvIntPtr(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func csvFloatPtr(value *float64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatFloat(*value, 'f', -1, 64)
}

func csvJSON(value interface{}) string {
	bytes, err := stdjson.Marshal(value)
	if err != nil || string(bytes) == "null" {
		return ""
	}
	return string(bytes)
}

func csvWorkExperienceJSON(workExperience []studentModel.WorkExperience) string {
	type exportWorkExperience struct {
		StartDate    string `json:"startDate,omitempty"`
		EndDate      string `json:"endDate,omitempty"`
		Organisation string `json:"organisation,omitempty"`
		Location     string `json:"location,omitempty"`
		Position     string `json:"position,omitempty"`
		Details      string `json:"details,omitempty"`
		IsVerified   bool   `json:"isVerified"`
	}

	exported := make([]exportWorkExperience, 0, len(workExperience))
	for _, experience := range workExperience {
		exported = append(exported, exportWorkExperience{
			StartDate:    csvDateTime(experience.StartDate),
			EndDate:      csvDateTime(experience.EndDate),
			Organisation: experience.Organisation,
			Location:     experience.Location,
			Position:     experience.Position,
			Details:      experience.Details,
			IsVerified:   experience.Verification.IsVerified,
		})
	}

	return csvJSON(exported)
}

func GetStudentRoleObjectID() primitive.ObjectID {
	if objID, err := primitive.ObjectIDFromHex(os.Getenv(constants.ENV_STUDENT_GROUP_OBJ_ID)); err != nil {
		return primitive.NilObjectID
	} else {
		return objID
	}
}

func (h *Handler) GetAllStudents(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)

	query := strings.TrimSpace(ctx.Query("query"))
	if len(query) < controller.MinStudentSearchQueryLength {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "query param is required (min 2 characters) — search by name or roll number",
		})
		return
	}

	startYear, err := strconv.Atoi(ctx.DefaultQuery("startYear", "0"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid startYear"})
		return
	}

	endYear, err := strconv.Atoi(ctx.DefaultQuery("endYear", "0"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid endYear"})
		return
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", strconv.Itoa(controller.DefaultStudentSearchLimit)))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}

	students, err := controller.SearchStudents(h.MongikClient, controller.StudentSearchFilter{
		Query:      query,
		StartYear:  startYear,
		EndYear:    endYear,
		Course:     ctx.Query("course"),
		Department: ctx.Query("department"),
		Limit:      limit,
	}, noCache)

	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "query must be at least") {
			status = http.StatusBadRequest
		}
		ctx.AbortWithStatusJSON(status, gin.H{
			"data":  nil,
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data":  students,
		"error": nil,
	})
}

func (h *Handler) GetStudentDirectory(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)

	value, exists := ctx.Get(constants.SESSION)
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Cannot get student",
		})
		return
	}

	currentStudent := value.(*model.StudentPopulated)

	batch := ctx.Query("batch")
	department := ctx.Query("department")
	course := ctx.Query("course")
	fields := ctx.Query("fields")

	students, err := controller.GetStudentDirectory(h.MongikClient, currentStudent, batch, department, course, fields, noCache)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"data":  nil,
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
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
		controller.InvalidateStudentCache(h.MongikClient, currentStudent.InstituteEmail)
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

func (h *Handler) HandlerGetStudentProfileById(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)
	studentId, err := primitive.ObjectIDFromHex(ctx.GetHeader("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	student, err := controller.GetStudentById(h.MongikClient, studentId, noCache)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	studentProfile := interfaces.StudentProfile{}
	controller.MapStudentToStudentProfile(&studentProfile, &student.Student, true)

	ctx.JSON(200, gin.H{
		"profile":             studentProfile,
		"isAcademicsVerified": student.Academics.Verification.IsVerified,
		"student": gin.H{
			"_id":              student.Id,
			"firstName":        student.FirstName,
			"middleName":       student.MiddleName,
			"lastName":         student.LastName,
			"email":            student.InstituteEmail,
			"rollNo":           student.RollNo,
			"department":       student.Department,
			"companiesAlloted": student.CompaniesAlloted,
			"isInterned":       student.IsInterned,
			"internCompany":    student.InternCompany,
			"hasPPO":           student.HasPPO,
			"ppoCompany":       student.PPOCompany,
			"isPlaced":         student.IsPlaced,
			"placedCompany":    student.PlacedCompany,
		},
	})
}

func (h *Handler) HandlerVerifyStudentProfile(ctx *gin.Context) {
	studentId, err := primitive.ObjectIDFromHex(ctx.GetHeader("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	// Get the admin who is verifying
	admin, exists := ctx.Get(constants.SESSION)
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	adminStudent := admin.(*model.StudentPopulated)

	student, err := controller.VerifyStudentProfile(h.MongikClient, studentId, adminStudent.Id)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	lastNameStr := ""
	if student.LastName != nil {
		lastNameStr = *student.LastName
	}
	if util.CheckRoleExists(&adminStudent.GroupDetails, constants.ROLE_ADMIN) {
		h.LogActivityDirect(adminStudent.Id, "EDIT", fmt.Sprintf("Verified student profile for %s %s (%s) - Roll No: %d", student.FirstName, lastNameStr, student.InstituteEmail, student.RollNo))
	}

	ctx.JSON(200, gin.H{"message": "Profile verified successfully", "student": student})
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
		controller.InvalidateStudentCache(h.MongikClient, currentStudent.InstituteEmail)
		ctx.JSON(200, gin.H{"student": updateResult})
	}

	ctx.JSON(200, gin.H{"student": currentStudent})
}

func (h *Handler) HandlerAdminUpdateStudentDetails(ctx *gin.Context) {
	studentId, err := primitive.ObjectIDFromHex(ctx.GetHeader("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	admin, exists := ctx.Get(constants.SESSION)
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	adminStudent := admin.(*model.StudentPopulated)

	var studentProfile interfaces.StudentProfile
	if errBinding := ctx.ShouldBindJSON(&studentProfile); errBinding != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": errBinding.Error()})
		return
	}

	studentCollection := h.MongikClient.MongoClient.Database(constants.DB).Collection(constants.COLLECTION_STUDENT)
	filter := bson.M{"_id": studentId}

	var currentStudent studentModel.Student
	if errFind := studentCollection.FindOne(ctx, filter).Decode(&currentStudent); errFind != nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Student not found"})
		return
	}

	// Admin update: map the provided profile directly onto the current student without restrictions
	controller.MapStudentToStudentProfile(&studentProfile, &currentStudent, false)
	currentStudent.UpdatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())

	if updateResult, errUpdate := db.ReplaceOne(h.MongikClient, constants.DB, constants.COLLECTION_STUDENT, filter, &currentStudent); errUpdate != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": errUpdate.Error()})
		return
	} else {
		controller.InvalidateStudentCache(h.MongikClient, currentStudent.InstituteEmail)
		lastNameStr := ""
		if currentStudent.LastName != nil {
			lastNameStr = *currentStudent.LastName
		}
		h.LogActivityDirect(adminStudent.Id, "EDIT", fmt.Sprintf("Updated student details for %s %s (%s) - Roll No: %d", currentStudent.FirstName, lastNameStr, currentStudent.InstituteEmail, currentStudent.RollNo))
		ctx.JSON(200, gin.H{"student": updateResult})
	}
}

func (h *Handler) HandlerUnverifyStudentProfilesByBatch(ctx *gin.Context) {
	var req interfaces.UnverifyBatchRequest
	if errBinding := ctx.ShouldBindJSON(&req); errBinding != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errBinding.Error()})
		return
	}

	if req.StartYear == 0 || req.EndYear == 0 || req.EndYear < req.StartYear {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid batch years"})
		return
	}

	admin, exists := ctx.Get(constants.SESSION)
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	adminStudent := admin.(*model.StudentPopulated)

	updatedCount, errs := controller.UnverifyStudentProfilesByBatch(h.MongikClient, req.StartYear, req.EndYear)

	if len(errs) > 0 {
		ctx.JSON(http.StatusPartialContent, gin.H{
			"message":      "Batch unverify completed with errors",
			"updatedCount": updatedCount,
			"errors":       errs,
		})
		return
	}

	h.LogActivityDirect(adminStudent.Id, "EDIT", fmt.Sprintf("Unverified profiles for batch: %d-%d (total %d students updated)", req.StartYear, req.EndYear, updatedCount))

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Successfully unverified all profiles in batch",
		"updatedCount": updatedCount,
	})
}

func (h *Handler) HandlerAdminUpdateStudentPlacementStatus(ctx *gin.Context) {
	studentId, err := primitive.ObjectIDFromHex(ctx.GetHeader("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	admin, exists := ctx.Get(constants.SESSION)
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	adminStudent := admin.(*model.StudentPopulated)

	var req interfaces.StudentPlacementStatusUpdate
	if errBinding := ctx.ShouldBindJSON(&req); errBinding != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errBinding.Error()})
		return
	}

	update := bson.M{}
	if req.IsInterned != nil {
		update["isInterned"] = *req.IsInterned
	}
	if req.InternCompany != nil {
		update["internCompany"] = strings.TrimSpace(*req.InternCompany)
	}
	if req.HasPPO != nil {
		update["hasPPO"] = *req.HasPPO
	}
	if req.PPOCompany != nil {
		update["ppoCompany"] = strings.TrimSpace(*req.PPOCompany)
	}
	if req.IsPlaced != nil {
		update["isPlaced"] = *req.IsPlaced
	}
	if req.PlacedCompany != nil {
		update["placedCompany"] = strings.TrimSpace(*req.PlacedCompany)
	}

	student, err := controller.UpdateStudentPlacementStatus(h.MongikClient, studentId, update)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	lastNameStr := ""
	if student.LastName != nil {
		lastNameStr = *student.LastName
	}
	h.LogActivityDirect(adminStudent.Id, "EDIT", fmt.Sprintf("Updated placement status for student %s %s (%s) - Roll No: %d", student.FirstName, lastNameStr, student.InstituteEmail, student.RollNo))

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Student placement status updated successfully",
		"student": student,
	})
}

func (h *Handler) HandlerAdminExportStudentsCSV(ctx *gin.Context) {
	startYear, err := strconv.Atoi(ctx.DefaultQuery("startYear", "0"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid startYear"})
		return
	}

	endYear, err := strconv.Atoi(ctx.DefaultQuery("endYear", "0"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid endYear"})
		return
	}

	status := ctx.Query("status")
	students, err := controller.GetStudentsForExport(h.MongikClient, startYear, endYear, status)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)
	headers := []string{
		"Roll Number",
		"First Name",
		"Middle Name",
		"Last Name",
		"Full Name",
		"Institute Email",
		"Personal Email",
		"Mobile",
		"Gender",
		"DOB",
		"Department",
		"Course",
		"Specialisation",
		"Batch Start Year",
		"Batch End Year",
		"Profile Picture Name",
		"Profile Picture MIME Type",
		"Profile Picture URL",
		"Profile Picture Is External",
		"Permanent Address",
		"Present Address",
		"Category",
		"Is PWD",
		"Is EWS",
		"Mother Tongue",
		"Father Name",
		"Father Occupation",
		"Mother Name",
		"Mother Occupation",
		"JEE Rank",
		"JEE Rank Category",
		"JEE Rank Is PWD",
		"JEE Rank Is EWS",
		"GATE Rank",
		"GATE Rank Category",
		"GATE Rank Is PWD",
		"GATE Rank Is EWS",
		"Class X Certification",
		"Class X Institute",
		"Class X Year",
		"Class X Score",
		"Class XII Certification",
		"Class XII Institute",
		"Class XII Year",
		"Class XII Score",
		"Undergraduate Certification",
		"Undergraduate Institute",
		"Undergraduate Year",
		"Undergraduate Score",
		"Postgraduate Certification",
		"Postgraduate Institute",
		"Postgraduate Year",
		"Postgraduate Score",
		"Diploma Certification",
		"Diploma Institute",
		"Diploma Year",
		"Diploma Score",
		"Honours",
		"Thesis End Date",
		"Education Gap",
		"Semester SPI 1",
		"Semester SPI 2",
		"Semester SPI 3",
		"Semester SPI 4",
		"Semester SPI 5",
		"Semester SPI 6",
		"Semester SPI 7",
		"Semester SPI 8",
		"Semester SPI 9",
		"Semester SPI 10",
		"Summer Term SPI 1",
		"Summer Term SPI 2",
		"Summer Term SPI 3",
		"Summer Term SPI 4",
		"Summer Term SPI 5",
		"Current CGPA",
		"Active Backlogs",
		"Total Backlogs",
		"Academics Is Verified",
		"LinkedIn URL",
		"LinkedIn Username",
		"LinkedIn Is Verified",
		"Github URL",
		"Github Username",
		"Github Is Verified",
		"CodeChef URL",
		"CodeChef Username",
		"CodeChef Is Verified",
		"Codeforces URL",
		"Codeforces Username",
		"Codeforces Is Verified",
		"LeetCode URL",
		"LeetCode Username",
		"LeetCode Is Verified",
		"Kaggle URL",
		"Kaggle Username",
		"Kaggle Is Verified",
		"Google Scholar URL",
		"Google Scholar Username",
		"Google Scholar Is Verified",
		"Microsoft Teams URL",
		"Microsoft Teams Username",
		"Microsoft Teams Is Verified",
		"Skype URL",
		"Skype Username",
		"Skype Is Verified",
		"Work Experience",
		"Video Resume",
		"Extras Is Verified",
		"Companies Alloted",
		"Interned",
		"Intern Company",
		"PPO",
		"PPO Company",
		"Placed",
		"Placed Company",
	}
	if err := writer.Write(headers); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, student := range *students {
		batchStartYear := ""
		batchEndYear := ""
		if student.Batch != nil {
			batchStartYear = strconv.Itoa(student.Batch.StartYear)
			batchEndYear = strconv.Itoa(student.Batch.EndYear)
		}

		nameParts := []string{student.FirstName}
		if student.MiddleName != nil && strings.TrimSpace(*student.MiddleName) != "" {
			nameParts = append(nameParts, strings.TrimSpace(*student.MiddleName))
		}
		if student.LastName != nil && strings.TrimSpace(*student.LastName) != "" {
			nameParts = append(nameParts, strings.TrimSpace(*student.LastName))
		}

		course := ""
		if student.Course != nil {
			course = string(*student.Course)
		}

		gender := ""
		if student.Gender != nil {
			gender = string(*student.Gender)
		}

		profilePictureName := ""
		profilePictureMimeType := ""
		profilePictureURL := ""
		profilePictureIsExternal := ""
		if student.ProfilePicture != nil {
			profilePictureName = student.ProfilePicture.Name
			profilePictureMimeType = student.ProfilePicture.MimeType
			profilePictureURL = student.ProfilePicture.URL
			profilePictureIsExternal = csvBoolText(student.ProfilePicture.IsExternal)
		}

		category := ""
		isPWD := ""
		isEWS := ""
		if student.Category != nil {
			category = student.Category.Category
			isPWD = csvBoolText(student.Category.IsPWD)
			isEWS = csvBoolText(student.Category.IsEWS)
		}

		rankFields := func(rank *studentModel.RankDetails) []string {
			if rank == nil {
				return []string{"", "", "", ""}
			}
			rankCategory := ""
			rankIsPWD := ""
			rankIsEWS := ""
			if rank.RankCategory != nil {
				rankCategory = rank.RankCategory.Category
				rankIsPWD = csvBoolText(rank.RankCategory.IsPWD)
				rankIsEWS = csvBoolText(rank.RankCategory.IsEWS)
			}
			return []string{strconv.Itoa(rank.Rank), rankCategory, rankIsPWD, rankIsEWS}
		}

		educationFields := func(education *studentModel.EducationDetails) []string {
			if education == nil {
				return []string{"", "", "", ""}
			}
			return []string{
				education.Certification,
				education.Institute,
				strconv.Itoa(education.Year),
				strconv.FormatFloat(education.Score, 'f', -1, 64),
			}
		}

		socialFields := func(social *studentModel.SocialProfile) []string {
			if social == nil {
				return []string{"", "", ""}
			}
			return []string{
				social.URL,
				social.Username,
				csvBoolText(social.Verification.IsVerified),
			}
		}

		row := []string{
			strconv.Itoa(student.RollNo),
			student.FirstName,
			csvStringPtr(student.MiddleName),
			csvStringPtr(student.LastName),
			strings.Join(nameParts, " "),
			student.InstituteEmail,
			student.PersonalEmail,
			student.Mobile,
			gender,
			csvDateTimePtr(student.DOB),
			student.Department,
			course,
			csvStringPtr(student.Specialisation),
			batchStartYear,
			batchEndYear,
			profilePictureName,
			profilePictureMimeType,
			profilePictureURL,
			profilePictureIsExternal,
			student.PermanentAddress,
			student.PresentAddress,
			category,
			isPWD,
			isEWS,
			student.MotherTongue,
			student.ParentsDetails.FatherName,
			student.ParentsDetails.FatherOccupation,
			student.ParentsDetails.MotherName,
			student.ParentsDetails.MotherOccupation,
		}

		row = append(row, rankFields(student.Academics.JEERank)...)
		row = append(row, rankFields(student.Academics.GATERank)...)
		row = append(row, educationFields(student.Academics.XthClass)...)
		row = append(row, educationFields(student.Academics.XIIthClass)...)
		row = append(row, educationFields(student.Academics.UnderGraduate)...)
		row = append(row, educationFields(student.Academics.PostGraduate)...)
		row = append(row, educationFields(student.Academics.Diploma)...)

		row = append(row,
			csvStringPtr(student.Academics.Honours),
			csvDateTimePtr(student.Academics.ThesisEndDate),
			csvIntPtr(student.Academics.EducationGap),
			csvFloatPtr(student.Academics.SemesterSPI.One),
			csvFloatPtr(student.Academics.SemesterSPI.Two),
			csvFloatPtr(student.Academics.SemesterSPI.Three),
			csvFloatPtr(student.Academics.SemesterSPI.Four),
			csvFloatPtr(student.Academics.SemesterSPI.Five),
			csvFloatPtr(student.Academics.SemesterSPI.Six),
			csvFloatPtr(student.Academics.SemesterSPI.Seven),
			csvFloatPtr(student.Academics.SemesterSPI.Eight),
			csvFloatPtr(student.Academics.SemesterSPI.Nine),
			csvFloatPtr(student.Academics.SemesterSPI.Ten),
			csvFloatPtr(student.Academics.SummerTermSPI.One),
			csvFloatPtr(student.Academics.SummerTermSPI.Two),
			csvFloatPtr(student.Academics.SummerTermSPI.Three),
			csvFloatPtr(student.Academics.SummerTermSPI.Four),
			csvFloatPtr(student.Academics.SummerTermSPI.Five),
			csvFloatPtr(student.Academics.CurrentCGPA),
			csvIntPtr(student.Academics.ActiveBacklogs),
			csvIntPtr(student.Academics.TotalBacklogs),
			csvBoolText(student.Academics.Verification.IsVerified),
		)

		row = append(row, socialFields(student.SocialProfiles.LinkedIn)...)
		row = append(row, socialFields(student.SocialProfiles.Github)...)
		row = append(row, socialFields(student.SocialProfiles.CodeChef)...)
		row = append(row, socialFields(student.SocialProfiles.Codeforces)...)
		row = append(row, socialFields(student.SocialProfiles.LeetCode)...)
		row = append(row, socialFields(student.SocialProfiles.Kaggle)...)
		row = append(row, socialFields(student.SocialProfiles.GoogleScholar)...)
		row = append(row, socialFields(student.SocialProfiles.MicrosoftTeams)...)
		row = append(row, socialFields(student.SocialProfiles.Skype)...)

		row = append(row,
			csvWorkExperienceJSON(student.WorkExperience),
			csvStringPtr(student.Extras.VideoResume),
			csvBoolText(student.Extras.Verification.IsVerified),
			strings.Join(student.CompaniesAlloted, "; "),
			csvBoolText(student.IsInterned),
			student.InternCompany,
			csvBoolText(student.HasPPO),
			student.PPOCompany,
			csvBoolText(student.IsPlaced),
			student.PlacedCompany,
		)
		if err := writer.Write(row); err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fileNameParts := []string{"students"}
	if startYear != 0 || endYear != 0 {
		fileNameParts = append(fileNameParts, strconv.Itoa(startYear)+"-"+strconv.Itoa(endYear))
	}
	if status != "" {
		fileNameParts = append(fileNameParts, strings.ToLower(status))
	}
	fileName := strings.Join(fileNameParts, "_") + ".csv"

	ctx.Header("Content-Type", "text/csv")
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.Data(http.StatusOK, "text/csv", buffer.Bytes())
}
