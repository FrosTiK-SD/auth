package controller

import (
	"sort"
	"strconv"
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

func CheckSocialProfile(updatedSocialProfile *studentModel.SocialProfile, currentSocialProfile *studentModel.SocialProfile) {
	if updatedSocialProfile == nil {
		currentSocialProfile = nil
		return
	}

	if updatedSocialProfile.URL != currentSocialProfile.URL || updatedSocialProfile.Username != currentSocialProfile.Username {
		currentSocialProfile.URL = updatedSocialProfile.URL
		currentSocialProfile.Username = updatedSocialProfile.Username
		SetVerificationToNotVerified(&currentSocialProfile.Verification)
	}
}

func InvalidateVerifiedFieldsOnChange(updated *studentModel.Student, current *studentModel.Student) {
	// invalidate academic details
	if !cmp.Equal(updated.Academics, current.Academics) {
		current.Academics = updated.Academics
		SetVerificationToNotVerified(&current.Academics.Verification)
	}

	// invalidate social profiles
	CheckSocialProfile(updated.SocialProfiles.LinkedIn, current.SocialProfiles.LinkedIn)
	CheckSocialProfile(updated.SocialProfiles.Github, current.SocialProfiles.Github)
	CheckSocialProfile(updated.SocialProfiles.MicrosoftTeams, current.SocialProfiles.MicrosoftTeams)
	CheckSocialProfile(updated.SocialProfiles.Skype, current.SocialProfiles.Skype)
	CheckSocialProfile(updated.SocialProfiles.GoogleScholar, current.SocialProfiles.GoogleScholar)
	CheckSocialProfile(updated.SocialProfiles.Codeforces, current.SocialProfiles.Codeforces)
	CheckSocialProfile(updated.SocialProfiles.CodeChef, current.SocialProfiles.CodeChef)
	CheckSocialProfile(updated.SocialProfiles.LeetCode, current.SocialProfiles.LeetCode)
	CheckSocialProfile(updated.SocialProfiles.Kaggle, current.SocialProfiles.Kaggle)

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

func GetAllStudents(mongikClient *models.Mongik, noCache bool, currentPage int, studentsPerPage int, search string) (*[]model.StudentPopulated, int, error) {
	var pipeline []bson.M

	// Add search functionality
	if search != "" && len(search) > 0 {
		var searchConditions []bson.M

		if search[0] >= '0' && search[0] <= '9' {
			// If search starts with a digit, search ONLY on rollNo
			// Validate that the search string is a valid number
			if _, err := strconv.Atoi(search); err != nil {
				return nil, 0, err
			}

			// Create lower and upper bounds for roll number range search
			lowerBound, upperBound := getRollNumberBounds(search)

			searchConditions = []bson.M{
				{
					"rollNo": bson.M{
						"$gte": lowerBound,
						"$lte": upperBound,
					},
				},
			}
		} else {
			// If search starts with a letter, search ONLY on firstName or lastName
			searchConditions = []bson.M{
				{"firstName": bson.M{"$regex": search, "$options": "i"}},
				{"lastName": bson.M{"$regex": search, "$options": "i"}},
			}
		}

		pipeline = append(pipeline, bson.M{
			"$match": bson.M{
				"$or": searchConditions,
			},
		})
	}

	// Count matching students before pagination
	countPipeline := make([]bson.M, len(pipeline))
	copy(countPipeline, pipeline)
	countPipeline = append(countPipeline, bson.M{"$count": "total"})

	countResult, err := db.Aggregate[map[string]int](mongikClient, constants.DB, constants.COLLECTION_STUDENT, countPipeline, noCache)
	if err != nil {
		return nil, 0, err
	}

	totalStudents := 0
	if len(countResult) > 0 {
		totalStudents = countResult[0]["total"]
	}

	pipeline = append(pipeline, bson.M{
		"$lookup": bson.M{
			"from":         constants.COLLECTION_GROUP,
			"localField":   "groups",
			"foreignField": "_id",
			"as":           "groups",
		},
	})

	if currentPage != 0 && studentsPerPage != 0 {
		skip := (currentPage - 1) * studentsPerPage
		pipeline = append(pipeline,
			bson.M{"$skip": skip},
			bson.M{"$limit": studentsPerPage},
		)
	}

	students, err := db.Aggregate[model.StudentPopulated](mongikClient, constants.DB, constants.COLLECTION_STUDENT, pipeline, noCache)
	if err != nil {
		return nil, 0, err
	}

	return &students, totalStudents, nil
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
