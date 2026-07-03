package controller

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/model"
	"github.com/FrosTiK-SD/auth/util"
	"github.com/FrosTiK-SD/models/misc"
	studentModel "github.com/FrosTiK-SD/models/student"
	db "github.com/FrosTiK-SD/mongik/db"
	models "github.com/FrosTiK-SD/mongik/models"
	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getAliasEmailList(email string) []string {
	var aliasEmailList []string
	aliasEmailList = append(aliasEmailList, email)
	aliasEmailList = append(aliasEmailList, strings.ReplaceAll(email, "iitbhu.ac.in", "itbhu.ac.in"))
	aliasEmailList = append(aliasEmailList, strings.ReplaceAll(email, "itbhu.ac.in", "iitbhu.ac.in"))
	sort.Strings(aliasEmailList)
	return aliasEmailList
}

func GetUserByEmail(mongikClient *models.Mongik, email *string, role *string, noCache bool) (*model.StudentPopulated, *string) {
	var studentPopulated model.StudentPopulated

	// Gets the alias emails
	emailList := getAliasEmailList(*email)

	// Query to DB
	studentPopulated, _ = db.AggregateOne[model.StudentPopulated](mongikClient, constants.DB, constants.COLLECTION_STUDENT, []bson.M{{
		"$match": bson.M{"email": bson.M{"$in": emailList}},
	}, {
		"$lookup": bson.M{
			"from":         constants.COLLECTION_GROUP,
			"localField":   "groups",
			"foreignField": "_id",
			"as":           "groups",
		},
	}}, noCache)

	// Now check if it is actually a student by the ROLES
	if !util.CheckRoleExists(&studentPopulated.GroupDetails, *role) {
		return nil, &constants.ERROR_NOT_A_STUDENT
	}

	return &studentPopulated, nil
}

func AssignUnVerifiedFields(updated *studentModel.Student, current *studentModel.Student) {
	// cannot change: groups, companies, batch, email, department, academics.verification, socialProfile.verification, metadata
	current.RollNo = updated.RollNo
	current.Course = updated.Course
	current.Specialisation = updated.Specialisation
	current.FirstName = updated.FirstName
	current.MiddleName = updated.MiddleName
	current.LastName = updated.LastName
	current.ProfilePicture = updated.ProfilePicture
	current.Gender = updated.Gender
	current.DOB = updated.DOB
	current.PermanentAddress = updated.PermanentAddress
	current.PresentAddress = updated.PresentAddress
	current.PersonalEmail = updated.PersonalEmail
	current.Mobile = updated.Mobile
	current.Category = updated.Category
	current.MotherTongue = updated.MotherTongue
	current.ParentsDetails = updated.ParentsDetails
	current.UpdatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
}

func SetVerificationToNotVerified(verification *misc.Verification) {
	verification.IsVerified = false
	verification.VerifiedBy = primitive.NilObjectID
	verification.VerifiedAt = 0
}

func SetVerificationToVerified(verification *misc.Verification, verifiedBy primitive.ObjectID) {
	verification.IsVerified = true
	verification.VerifiedBy = verifiedBy
	verification.VerifiedAt = primitive.NewDateTimeFromTime(time.Now())
}

func VerifyStudentProfile(mongikClient *models.Mongik, studentId primitive.ObjectID, verifiedBy primitive.ObjectID) (*studentModel.Student, error) {
	studentCollection := mongikClient.MongoClient.Database(constants.DB).Collection(constants.COLLECTION_STUDENT)

	var student studentModel.Student
	filter := bson.M{"_id": studentId}
	if err := studentCollection.FindOne(context.Background(), filter).Decode(&student); err != nil {
		return nil, err
	}

	// Verify academics
	SetVerificationToVerified(&student.Academics.Verification, verifiedBy)

	// Verify social profiles
	verifySocialProfile := func(sp *studentModel.SocialProfile) {
		if sp != nil {
			SetVerificationToVerified(&sp.Verification, verifiedBy)
		}
	}
	verifySocialProfile(student.SocialProfiles.LinkedIn)
	verifySocialProfile(student.SocialProfiles.Github)
	verifySocialProfile(student.SocialProfiles.MicrosoftTeams)
	verifySocialProfile(student.SocialProfiles.Skype)
	verifySocialProfile(student.SocialProfiles.GoogleScholar)
	verifySocialProfile(student.SocialProfiles.Codeforces)
	verifySocialProfile(student.SocialProfiles.CodeChef)
	verifySocialProfile(student.SocialProfiles.LeetCode)
	verifySocialProfile(student.SocialProfiles.Kaggle)

	// Verify work experience
	for i := range student.WorkExperience {
		SetVerificationToVerified(&student.WorkExperience[i].Verification, verifiedBy)
	}

	// Verify extras
	SetVerificationToVerified(&student.Extras.Verification, verifiedBy)

	// Save
	if _, err := db.ReplaceOne(mongikClient, constants.DB, constants.COLLECTION_STUDENT, filter, &student); err != nil {
		return nil, err
	}
	InvalidateStudentCache(mongikClient, student.InstituteEmail)
	return &student, nil
}

func UnverifyStudentProfilesByBatch(mongikClient *models.Mongik, startYear int, endYear int) (int, []error) {
	studentCollection := mongikClient.MongoClient.Database(constants.DB).Collection(constants.COLLECTION_STUDENT)

	filter := bson.M{
		"batch.startYear": startYear,
		"batch.endYear":   endYear,
	}

	cursor, err := studentCollection.Find(context.Background(), filter)
	if err != nil {
		return 0, []error{err}
	}
	defer cursor.Close(context.Background())

	var errors []error
	updatedCount := 0

	for cursor.Next(context.Background()) {
		var student studentModel.Student
		if err := cursor.Decode(&student); err != nil {
			errors = append(errors, err)
			continue
		}

		SetVerificationToNotVerified(&student.Academics.Verification)

		unverifySocialProfile := func(sp *studentModel.SocialProfile) {
			if sp != nil {
				SetVerificationToNotVerified(&sp.Verification)
			}
		}
		unverifySocialProfile(student.SocialProfiles.LinkedIn)
		unverifySocialProfile(student.SocialProfiles.Github)
		unverifySocialProfile(student.SocialProfiles.MicrosoftTeams)
		unverifySocialProfile(student.SocialProfiles.Skype)
		unverifySocialProfile(student.SocialProfiles.GoogleScholar)
		unverifySocialProfile(student.SocialProfiles.Codeforces)
		unverifySocialProfile(student.SocialProfiles.CodeChef)
		unverifySocialProfile(student.SocialProfiles.LeetCode)
		unverifySocialProfile(student.SocialProfiles.Kaggle)

		for i := range student.WorkExperience {
			SetVerificationToNotVerified(&student.WorkExperience[i].Verification)
		}

		SetVerificationToNotVerified(&student.Extras.Verification)
		student.UpdatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())

		_, updateErr := studentCollection.ReplaceOne(
			context.Background(),
			bson.M{"_id": student.Id},
			&student,
		)
		if updateErr != nil {
			errors = append(errors, updateErr)
			continue
		}
		InvalidateStudentCache(mongikClient, student.InstituteEmail)
		updatedCount++
	}
	// clear cache for the entire batch
	db.DBCacheReset(mongikClient, constants.COLLECTION_STUDENT)

	return updatedCount, errors
}

func CheckSocialProfile(updatedSocialProfile *studentModel.SocialProfile, currentSocialProfile **studentModel.SocialProfile) {
	if updatedSocialProfile == nil {
		*currentSocialProfile = nil
		return
	}

	if *currentSocialProfile == nil {
		*currentSocialProfile = new(studentModel.SocialProfile)
		(*currentSocialProfile).URL = updatedSocialProfile.URL
		(*currentSocialProfile).Username = updatedSocialProfile.Username
		SetVerificationToNotVerified(&(*currentSocialProfile).Verification)
		return
	}

	if updatedSocialProfile.URL != (*currentSocialProfile).URL || updatedSocialProfile.Username != (*currentSocialProfile).Username {
		(*currentSocialProfile).URL = updatedSocialProfile.URL
		(*currentSocialProfile).Username = updatedSocialProfile.Username
		SetVerificationToNotVerified(&(*currentSocialProfile).Verification)
	}
}

func InvalidateVerifiedFieldsOnChange(updated *studentModel.Student, current *studentModel.Student) {
	// invalidate academic details
	if !cmp.Equal(updated.Academics, current.Academics) {
		current.Academics = updated.Academics
		SetVerificationToNotVerified(&current.Academics.Verification)
	}

	// invalidate social profiles
	CheckSocialProfile(updated.SocialProfiles.LinkedIn, &current.SocialProfiles.LinkedIn)
	CheckSocialProfile(updated.SocialProfiles.Github, &current.SocialProfiles.Github)
	CheckSocialProfile(updated.SocialProfiles.MicrosoftTeams, &current.SocialProfiles.MicrosoftTeams)
	CheckSocialProfile(updated.SocialProfiles.Skype, &current.SocialProfiles.Skype)
	CheckSocialProfile(updated.SocialProfiles.GoogleScholar, &current.SocialProfiles.GoogleScholar)
	CheckSocialProfile(updated.SocialProfiles.Codeforces, &current.SocialProfiles.Codeforces)
	CheckSocialProfile(updated.SocialProfiles.CodeChef, &current.SocialProfiles.CodeChef)
	CheckSocialProfile(updated.SocialProfiles.LeetCode, &current.SocialProfiles.LeetCode)
	CheckSocialProfile(updated.SocialProfiles.Kaggle, &current.SocialProfiles.Kaggle)

	var newWorkExperienceArray []studentModel.WorkExperience

	// invalidate work experience
	for _, updatedWorkExp := range updated.WorkExperience {
		isUpdated := true
		for _, currentWorkExp := range current.WorkExperience {
			if cmp.Equal(updatedWorkExp, currentWorkExp) {
				newWorkExperienceArray = append(newWorkExperienceArray, currentWorkExp)
				isUpdated = false
				break
			}
		}

		if isUpdated {
			SetVerificationToNotVerified(&updatedWorkExp.Verification)
			newWorkExperienceArray = append(newWorkExperienceArray, updatedWorkExp)
		}
	}

	current.WorkExperience = newWorkExperienceArray

	// invalidate extra
	if !cmp.Equal(updated.Extras, current.Extras) {
		updated.Extras = current.Extras
		SetVerificationToNotVerified(&updated.Extras.Verification)
	}
}

func GetAllStudents(mongikClient *models.Mongik, noCache bool) (*[]model.StudentPopulated, error) {
	students, err := db.Aggregate[model.StudentPopulated](mongikClient, constants.DB, constants.COLLECTION_STUDENT, []bson.M{
		{
			"$lookup": bson.M{
				"from":         constants.COLLECTION_GROUP,
				"localField":   "groups",
				"foreignField": "_id",
				"as":           "groups",
			},
		},
	}, noCache)
	return &students, err
}

func BuildStudentExportFilter(startYear int, endYear int, status string) bson.M {
	filter := bson.M{}

	if startYear != 0 {
		filter["batch.startYear"] = startYear
	}
	if endYear != 0 {
		filter["batch.endYear"] = endYear
	}

	switch strings.ToLower(strings.TrimSpace(status)) {
	case "placed":
		filter["isPlaced"] = true
	case "ppo":
		filter["hasPPO"] = true
	case "intern", "interned", "internship":
		filter["isInterned"] = true
	case "allowed", "alloted", "allotted":
		filter["companiesAlloted.0"] = bson.M{"$exists": true}
	case "unplaced":
		filter["isPlaced"] = bson.M{"$ne": true}
	case "not-ppo":
		filter["hasPPO"] = bson.M{"$ne": true}
	case "not-interned":
		filter["isInterned"] = bson.M{"$ne": true}
	}

	return filter
}

func GetStudentsForExport(mongikClient *models.Mongik, startYear int, endYear int, status string) (*[]studentModel.Student, error) {
	studentCollection := mongikClient.MongoClient.Database(constants.DB).Collection(constants.COLLECTION_STUDENT)
	filter := BuildStudentExportFilter(startYear, endYear, status)

	cursor, err := studentCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var students []studentModel.Student
	if err := cursor.All(context.Background(), &students); err != nil {
		return nil, err
	}

	return &students, nil
}

func UpdateStudentPlacementStatus(mongikClient *models.Mongik, studentId primitive.ObjectID, update bson.M) (*studentModel.Student, error) {
	studentCollection := mongikClient.MongoClient.Database(constants.DB).Collection(constants.COLLECTION_STUDENT)
	filter := bson.M{"_id": studentId}

	update["updatedAt"] = primitive.NewDateTimeFromTime(time.Now().UTC())

	var currentStudent studentModel.Student
	if err := studentCollection.FindOne(context.Background(), filter).Decode(&currentStudent); err != nil {
		return nil, err
	}

	if value, ok := update["isInterned"].(bool); ok {
		currentStudent.IsInterned = value
	}
	if value, ok := update["internCompany"].(string); ok {
		currentStudent.InternCompany = value
	}
	if value, ok := update["hasPPO"].(bool); ok {
		currentStudent.HasPPO = value
	}
	if value, ok := update["ppoCompany"].(string); ok {
		currentStudent.PPOCompany = value
	}
	if value, ok := update["isPlaced"].(bool); ok {
		currentStudent.IsPlaced = value
	}
	if value, ok := update["placedCompany"].(string); ok {
		currentStudent.PlacedCompany = value
	}
	currentStudent.UpdatedAt = update["updatedAt"].(primitive.DateTime)

	if _, err := db.ReplaceOne(mongikClient, constants.DB, constants.COLLECTION_STUDENT, filter, &currentStudent); err != nil {
		return nil, err
	}
	InvalidateStudentCache(mongikClient, currentStudent.InstituteEmail)

	return &currentStudent, nil
}

func GetStudentById(mongikClient *models.Mongik, _id primitive.ObjectID, noCache bool) (*model.StudentPopulated, error) {
	student, err := db.AggregateOne[model.StudentPopulated](mongikClient, constants.DB, constants.COLLECTION_STUDENT, []bson.M{
		{
			"$match": bson.M{
				"_id": _id,
			},
		},
		{
			"$lookup": bson.M{
				"from":         constants.COLLECTION_GROUP,
				"localField":   "groups",
				"foreignField": "_id",
				"as":           "groups",
			},
		},
	},
		noCache)
	return &student, err
}

func GetStudentDirectory(mongikClient *models.Mongik, currentStudent *model.StudentPopulated, batch string, department string, course string, fields string, noCache bool) (*[]model.StudentPopulated, error) {
	matchFilter := bson.M{}

	if batch != "" {
		matchFilter["batch"] = bson.M{
			"startYear": currentStudent.Batch.StartYear,
			"endYear":   currentStudent.Batch.EndYear,
		}
	} else {
		matchFilter["batch"] = bson.M{
			"startYear": currentStudent.Batch.StartYear,
			"endYear":   currentStudent.Batch.EndYear,
		}
	}

	if department != "" {
		matchFilter["department"] = department
	} else {
		matchFilter["department"] = currentStudent.Department
	}

	if course != "" {
		matchFilter["course"] = course
	}

	pipeline := []bson.M{
		{
			"$match": matchFilter,
		},
	}

	if fields == "basic" {
		pipeline = append(pipeline, bson.M{
			"$project": bson.M{
				"firstName": 1,
				"lastName":  1,
				"rollNo":    1,
			},
		})
	} else {
		pipeline = append(pipeline, bson.M{
			"$lookup": bson.M{
				"from":         constants.COLLECTION_GROUP,
				"localField":   "groups",
				"foreignField": "_id",
				"as":           "groups",
			},
		})
	}

	pipeline = append(pipeline, bson.M{
		"$limit": 500,
	})

	students, err := db.Aggregate[model.StudentPopulated](mongikClient, constants.DB, constants.COLLECTION_STUDENT, pipeline, noCache)

	return &students, err
}

func GetAllStudentsOfRole(mongikClient *models.Mongik, role string, noCache bool) (*[]model.StudentPopulated, error) {
	roleStudents, err := db.Aggregate[model.StudentPopulated](mongikClient, constants.DB, constants.COLLECTION_STUDENT, []bson.M{
		{
			"$lookup": bson.M{
				"from":         constants.COLLECTION_GROUP,
				"localField":   "groups",
				"foreignField": "_id",
				"as":           "groups",
			},
		},
		{
			"$match": bson.M{
				"groups": bson.M{
					"$elemMatch": bson.M{
						"roles": role,
					},
				},
			},
		},
	},
		noCache)

	return &roleStudents, err
}

// delete student profile cache key from Redis/BigCache
func InvalidateStudentCache(mongikClient *models.Mongik, email string) {
	emailList := getAliasEmailList(email)
	key := fmt.Sprintf("students | DB_AGGREGATEONE | [map[$match:map[email:map[$in:%v]]] map[$lookup:map[as:groups foreignField:_id from:groups localField:groups]]] | ", emailList)

	if mongikClient.RedisClient != nil {
		mongikClient.RedisClient.Del(context.Background(), key)
	}
	if mongikClient.CacheClient != nil {
		mongikClient.CacheClient.Delete(key)
	}
}
