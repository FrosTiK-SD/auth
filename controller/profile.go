package controller

import (
	"fmt"
	"reflect"

	"strconv"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/interfaces"
	"github.com/FrosTiK-SD/auth/model"
	constantModel "github.com/FrosTiK-SD/models/constant"
	studentModel "github.com/FrosTiK-SD/models/student"
	"github.com/modern-go/reflect2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var PointerToNilString *string = nil
var PointerToNilInteger *int = nil
var PointerToNilFloat64 *float64 = nil

func AssignReservationCategory(category *interfaces.GenericField, isEWS *interfaces.GenericField, isPWD *interfaces.GenericField, rc *studentModel.ReservationCategory, forward bool) {
	if reflect2.IsNil(rc) {
		category.IsNull = true
		isEWS.IsNull = true
		isPWD.IsNull = true
	} else {
		category.Value = rc.Category
		isEWS.Value = rc.IsEWS
		isPWD.Value = rc.IsPWD
	}

	category.DataType = constants.TYPE_STRING
	isEWS.DataType = constants.TYPE_BOOL
	isPWD.DataType = constants.TYPE_BOOL
}

func AssignSocialProfile(field *interfaces.GenericField, social *studentModel.SocialProfile, forward bool) {
	field.DataType = constants.TYPE_SOCIAL

	if social != nil {
		field.IsNull = reflect2.IsNil(field)

		if !field.IsNull {
			field.Value = social.URL + "|" + social.Username
		}

		return
	}

	field.IsNull = true
	field.Value = nil
}

func AssignNilPossibleValue[V int | float64 | string | constantModel.Course | constantModel.Gender | primitive.DateTime](field *interfaces.GenericField, value *V, forward bool) {
	field.Value = value
	field.IsNull = reflect2.IsNil(value)
	field.DataType = fmt.Sprintf("%v", reflect.TypeOf(value))

}

func AssignNotNilValue[V int | float64 | string | constantModel.Course](field *interfaces.GenericField, value *V, forward bool) {
	field.Value = *value
	field.DataType = fmt.Sprintf("%v", reflect.TypeOf(*value))
}

func AssignPastAcademics(field *interfaces.ProfilePastEducation, education *studentModel.EducationDetails, forward bool) {
	if education != nil {
		AssignNotNilValue(&field.Certification, &education.Certification, forward)
		AssignNotNilValue(&field.Institute, &education.Institute, forward)
		AssignNotNilValue(&field.Year, &education.Year, forward)
		AssignNotNilValue(&field.Score, &education.Score, forward)
		return
	}

	AssignNilPossibleValue(&field.Certification, PointerToNilString, forward)
	AssignNilPossibleValue(&field.Institute, PointerToNilString, forward)
	AssignNilPossibleValue(&field.Year, PointerToNilInteger, forward)
	AssignNilPossibleValue(&field.Score, PointerToNilFloat64, forward)
}

func AssignRankValue(field *interfaces.GenericRank, rankDetails *studentModel.RankDetails, forward bool) {
	if rankDetails != nil {
		AssignNotNilValue(&field.Rank, &rankDetails.Rank, forward)
		AssignReservationCategory(&field.Rank, &field.IsEWS, &field.IsPWD, &rankDetails.RankCategory, forward)
		return

	}

	AssignNilPossibleValue(&field.Rank, PointerToNilInteger, forward)
	AssignReservationCategory(&field.Category, &field.IsEWS, &field.IsPWD, nil, forward)
}

func MapProfilePersonal(profile *interfaces.ProfilePersonal, student *model.StudentPopulated, forward bool) {
	AssignNotNilValue(&profile.FirstName, &student.FirstName, forward)
	AssignNilPossibleValue(&profile.MiddleName, student.MiddleName, forward)
	AssignNilPossibleValue(&profile.LastName, student.LastName, forward)

	AssignNilPossibleValue(&profile.Gender, student.Gender, forward)
	AssignNilPossibleValue(&profile.DOB, student.DOB, forward)
	AssignNotNilValue(&profile.PermanentAddress, &student.PermanentAddress, forward)
	AssignNotNilValue(&profile.PresentAddress, &student.PresentAddress, forward)
	AssignNotNilValue(&profile.PersonalEmail, &student.PersonalEmail, forward)
	AssignNilPossibleValue(&profile.Mobile, student.Mobile, forward)
	AssignReservationCategory(&profile.Category, &profile.IsEWS, &profile.IsPWD, student.Category, forward)
	AssignNotNilValue(&profile.MotherTongue, &student.MotherTongue, forward)
	AssignNotNilValue(&profile.FatherName, &student.ParentsDetails.FatherName, forward)
	AssignNotNilValue(&profile.MotherName, &student.ParentsDetails.MotherName, forward)
	AssignNotNilValue(&profile.FatherOccupation, &student.ParentsDetails.FatherOccupation, forward)
	AssignNotNilValue(&profile.MotherOccupation, &student.ParentsDetails.MotherOccupation, forward)

	// required
	profile.FirstName.IsRequired = true
	profile.DOB.IsRequired = true
	profile.PermanentAddress.IsRequired = true
	profile.PersonalEmail.IsRequired = true
	profile.Mobile.IsRequired = true
}

func MapProfileCurrentAcademics(profile *interfaces.ProfileCurrentAcademics, academics *studentModel.Academics, forward bool) {
	AssignNilPossibleValue(&profile.SemesterSPI.One, academics.SemesterSPI.One, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Two, academics.SemesterSPI.Two, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Three, academics.SemesterSPI.Three, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Four, academics.SemesterSPI.Four, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Five, academics.SemesterSPI.Five, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Six, academics.SemesterSPI.Six, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Seven, academics.SemesterSPI.Seven, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Eight, academics.SemesterSPI.Eight, forward)

	AssignNilPossibleValue(&profile.SummerTermSPI.One, academics.SummerTermSPI.One, forward)
	AssignNilPossibleValue(&profile.SummerTermSPI.Two, academics.SummerTermSPI.Two, forward)
	AssignNilPossibleValue(&profile.SummerTermSPI.Three, academics.SummerTermSPI.Three, forward)
	AssignNilPossibleValue(&profile.SummerTermSPI.Four, academics.SummerTermSPI.Four, forward)
	AssignNilPossibleValue(&profile.SummerTermSPI.Five, academics.SummerTermSPI.Five, forward)
}

func AssignBatch(profile *interfaces.GenericField, institute *studentModel.Student, forward bool) {
	profile.IsNull = reflect2.IsNil(institute.Batch)
	profile.DataType = constants.TYPE_STRING
	if !profile.IsNull {
		profile.Value = strconv.Itoa(institute.Batch.StartYear) + "-" + strconv.Itoa(institute.Batch.EndYear)
	}
}

func MapProfileSocials(profile *interfaces.ProfileSocials, socials *studentModel.SocialProfiles, forward bool) {
	AssignSocialProfile(&profile.LinkedIn, socials.LinkedIn, forward)
	AssignSocialProfile(&profile.Github, socials.Github, forward)
	AssignSocialProfile(&profile.CodeChef, socials.CodeChef, forward)
	AssignSocialProfile(&profile.Codeforces, socials.Codeforces, forward)
	AssignSocialProfile(&profile.Leetcode, socials.LeetCode, forward)
	AssignSocialProfile(&profile.GoogleScholar, socials.GoogleScholar, forward)
	AssignSocialProfile(&profile.MicrosoftTeams, socials.MicrosoftTeams, forward)
	AssignSocialProfile(&profile.Kaggle, socials.Kaggle, forward)
	AssignSocialProfile(&profile.Skype, socials.Skype, forward)
}

func MapProfileInstitute(profile *interfaces.ProfileInstitute, institute *studentModel.Student, forward bool) {
	AssignBatch(&profile.Batch, institute, forward)
	AssignNotNilValue(&profile.RollNumber, &institute.RollNo, forward)
	AssignNotNilValue(&profile.InstituteEmail, &institute.InstituteEmail, forward)
	AssignNotNilValue(&profile.Department, &institute.Department, forward)
	AssignNilPossibleValue(&profile.EducationGap, institute.Academics.EducationGap, forward)
	AssignNotNilValue(&profile.Course, institute.Course, forward)
	AssignNilPossibleValue(&profile.Specialisation, institute.Specialisation, forward)
	AssignNilPossibleValue(&profile.Honours, institute.Academics.Honours, forward)
	AssignNilPossibleValue(&profile.ThesisEndDate, institute.Academics.ThesisEndDate, forward)
}

func MapPastAcademics(profile *interfaces.ProfilePastAcademics, institute *studentModel.Academics, forward bool) {
	AssignPastAcademics(&profile.ClassX, institute.XthClass, forward)
	AssignPastAcademics(&profile.ClassXII, institute.XIIthClass, forward)
	AssignPastAcademics(&profile.Undergraduate, institute.UnderGraduate, forward)
	AssignPastAcademics(&profile.Postgraduate, institute.PostGraduate, forward)
}

func MapRanks(profile *interfaces.ProfilePastAcademics, rank *studentModel.Academics, forward bool) {
	AssignRankValue(&profile.JeeRank, rank.JEERank, forward)
	AssignRankValue(&profile.GateRank, rank.GATERank, forward)
}

func MapStudentToStudentProfile(profile *interfaces.StudentProfile, student *model.StudentPopulated, forward bool) {
	// Profile
	MapProfilePersonal(&profile.Profile.PersonalProfile, student, forward)
	MapProfileSocials(&profile.Profile.SocialProfile, &student.SocialProfiles, forward)
	MapProfileInstitute(&profile.Profile.InstituteProfile, &student.Student, forward)

	// Past Academics
	MapPastAcademics(&profile.PastAcademics, &student.Academics, forward)
	MapRanks(&profile.PastAcademics, &student.Academics, forward)

	// Current Academics
	MapProfileCurrentAcademics(&profile.CurrentAcademics, &student.Academics, forward)
}
