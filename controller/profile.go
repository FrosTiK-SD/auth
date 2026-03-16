package controller

import (
	"fmt"
	"reflect"

	"strconv"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/interfaces"
	constantModel "github.com/FrosTiK-SD/models/constant"
	studentModel "github.com/FrosTiK-SD/models/student"
	"github.com/modern-go/reflect2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var PointerToNilString *string = nil
var PointerToNilInteger *int = nil
var PointerToNilFloat64 *float64 = nil
var PointerToNilReservationCategory *studentModel.ReservationCategory = nil
var GenderChoices = []constantModel.Gender{constantModel.MALE, constantModel.FEMALE, constantModel.OTHER_GENDER}
var ReservationCategoryChoices = []string{"GEN", "OBC", "OBC-NCL", "SC", "ST"}

func AssignReservationCategory(category *interfaces.GenericField[string], isEWS *interfaces.GenericField[bool], isPWD *interfaces.GenericField[bool], rc **studentModel.ReservationCategory, forward bool) {
	if forward {
		if reflect2.IsNil(*rc) {
			category.IsNull = true
			isEWS.IsNull = true
			isPWD.IsNull = true
		} else {
			category.Value = (**rc).Category
			isEWS.Value = (**rc).IsEWS
			isPWD.Value = (**rc).IsPWD
		}

		category.DataType = constants.TYPE_CHOICES
		category.DataChoices = &ReservationCategoryChoices
		isEWS.DataType = constants.TYPE_BOOL
		isPWD.DataType = constants.TYPE_BOOL

		return
	}

	// backward mapping

	if category.IsNull {
		return
	}

	*rc = new(studentModel.ReservationCategory)
	(**rc).Category = category.Value
	(**rc).IsEWS = isEWS.Value
	(**rc).IsPWD = isPWD.Value
}

func AssignSocialProfile(profile *interfaces.SocialProfile, social **studentModel.SocialProfile, forward bool) {
	if forward {
		profile.URL.DataType = constants.TYPE_STRING
		profile.Username.DataType = constants.TYPE_STRING

		if *social != nil {
			profile.URL.IsNull = false
			profile.Username.IsNull = false

			profile.URL.Value = (*social).URL
			profile.Username.Value = (*social).Username

			profile.URL.IsVerified = &(*social).Verification.IsVerified
			profile.Username.IsVerified = &(*social).Verification.IsVerified

			return
		}

		profile.URL.IsNull = true
		profile.Username.IsNull = true

		return
	}

	// backward mapping

	if profile.URL.IsNull || profile.Username.IsNull {
		return
	}

	*social = new(studentModel.SocialProfile)

	(**social).URL = profile.URL.Value
	(**social).Username = profile.Username.Value
}

func AssignNilPossibleValue[V int | float64 | string | constantModel.Course | constantModel.Gender | primitive.DateTime](field *interfaces.GenericField[V], value **V, forward bool) {
	if forward {
		field.IsNull = reflect2.IsNil(*value)
		if !field.IsNull {
			field.Value = **value
		}
		field.DataType = fmt.Sprintf("%v", reflect.TypeOf(*value))[1:]
		return
	}

	// backward mapping

	if field.IsNull {
		return
	}

	*value = new(V)
	**value = field.Value
}

func AssignNotNilValue[V int | float64 | string | constantModel.Course](field *interfaces.GenericField[V], value *V, forward bool) {
	if forward {
		field.Value = *value
		field.DataType = fmt.Sprintf("%v", reflect.TypeOf(*value))
		return
	}

	// backward mapping

	*value = field.Value
}

func AssignPastAcademics(field *interfaces.ProfilePastEducation, education **studentModel.EducationDetails, forward bool) {
	if forward {
		if *education != nil {
			AssignNotNilValue(&field.Certification, &(*education).Certification, forward)
			AssignNotNilValue(&field.Institute, &(*education).Institute, forward)
			AssignNotNilValue(&field.Year, &(*education).Year, forward)
			AssignNotNilValue(&field.Score, &(*education).Score, forward)
			return
		}

		AssignNilPossibleValue(&field.Certification, &PointerToNilString, forward)
		AssignNilPossibleValue(&field.Institute, &PointerToNilString, forward)
		AssignNilPossibleValue(&field.Year, &PointerToNilInteger, forward)
		AssignNilPossibleValue(&field.Score, &PointerToNilFloat64, forward)
		return
	}

	// backward

	if field.Certification.IsNull || field.Institute.IsNull {
		return
	}

	*education = new(studentModel.EducationDetails)
	AssignNotNilValue(&field.Certification, &(*education).Certification, forward)
	AssignNotNilValue(&field.Institute, &(*education).Institute, forward)
	AssignNotNilValue(&field.Year, &(*education).Year, forward)
	AssignNotNilValue(&field.Score, &(*education).Score, forward)

}

func AssignRankValue(field *interfaces.GenericRank, rankDetails **studentModel.RankDetails, forward bool) {
	if forward {
		if *rankDetails != nil {
			AssignNotNilValue(&field.Rank, &(**rankDetails).Rank, forward)
			AssignReservationCategory(&field.Category, &field.IsEWS, &field.IsPWD, &(**rankDetails).RankCategory, forward)
			return
		}

		AssignNilPossibleValue(&field.Rank, &PointerToNilInteger, forward)
		AssignReservationCategory(&field.Category, &field.IsEWS, &field.IsPWD, &PointerToNilReservationCategory, forward)
		return
	}

	// backward mapping

	if field.Rank.IsNull {
		return
	}

	*rankDetails = new(studentModel.RankDetails)
	AssignNotNilValue(&field.Rank, &(*rankDetails).Rank, forward)
	AssignReservationCategory(&field.Category, &field.IsEWS, &field.IsPWD, &(*rankDetails).RankCategory, forward)
}

func MapProfilePersonal(profile *interfaces.ProfilePersonal, student *studentModel.Student, forward bool) {
	AssignNotNilValue(&profile.FirstName, &student.FirstName, forward)
	AssignNilPossibleValue(&profile.MiddleName, &student.MiddleName, forward)
	AssignNilPossibleValue(&profile.LastName, &student.LastName, forward)

	AssignNilPossibleValue(&profile.Gender, &student.Gender, forward)
	AssignNilPossibleValue(&profile.DOB, &student.DOB, forward)
	AssignNotNilValue(&profile.PermanentAddress, &student.PermanentAddress, forward)
	AssignNotNilValue(&profile.PresentAddress, &student.PresentAddress, forward)
	AssignNotNilValue(&profile.PersonalEmail, &student.PersonalEmail, forward)
	AssignNotNilValue(&profile.Mobile, &student.Mobile, forward)
	AssignReservationCategory(&profile.Category, &profile.IsEWS, &profile.IsPWD, &student.Category, forward)
	AssignNotNilValue(&profile.MotherTongue, &student.MotherTongue, forward)
	AssignNotNilValue(&profile.FatherName, &student.ParentsDetails.FatherName, forward)
	AssignNotNilValue(&profile.MotherName, &student.ParentsDetails.MotherName, forward)
	AssignNotNilValue(&profile.FatherOccupation, &student.ParentsDetails.FatherOccupation, forward)
	AssignNotNilValue(&profile.MotherOccupation, &student.ParentsDetails.MotherOccupation, forward)

	if forward {
		// set required field
		profile.FirstName.IsRequired = true
		profile.DOB.IsRequired = true
		profile.PermanentAddress.IsRequired = true
		profile.PersonalEmail.IsRequired = true
		profile.Mobile.IsRequired = true

		// set choices
		profile.Gender.DataType = constants.TYPE_CHOICES
		profile.Gender.DataChoices = &GenderChoices
	}
}

func MapProfileCurrentAcademics(profile *interfaces.ProfileCurrentAcademics, academics *studentModel.Academics, forward bool) {
	AssignNilPossibleValue(&profile.SemesterSPI.One, &academics.SemesterSPI.One, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Two, &academics.SemesterSPI.Two, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Three, &academics.SemesterSPI.Three, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Four, &academics.SemesterSPI.Four, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Five, &academics.SemesterSPI.Five, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Six, &academics.SemesterSPI.Six, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Seven, &academics.SemesterSPI.Seven, forward)
	AssignNilPossibleValue(&profile.SemesterSPI.Eight, &academics.SemesterSPI.Eight, forward)

	AssignNilPossibleValue(&profile.SummerTermSPI.One, &academics.SummerTermSPI.One, forward)
	AssignNilPossibleValue(&profile.SummerTermSPI.Two, &academics.SummerTermSPI.Two, forward)
	AssignNilPossibleValue(&profile.SummerTermSPI.Three, &academics.SummerTermSPI.Three, forward)
	AssignNilPossibleValue(&profile.SummerTermSPI.Four, &academics.SummerTermSPI.Four, forward)
	AssignNilPossibleValue(&profile.SummerTermSPI.Five, &academics.SummerTermSPI.Five, forward)

	AssignNilPossibleValue(&profile.Misc.CurrentCGPA, &academics.CurrentCGPA, forward)
	AssignNilPossibleValue(&profile.Misc.ActiveBacklogs, &academics.ActiveBacklogs, forward)
	AssignNilPossibleValue(&profile.Misc.TotalBacklogs, &academics.TotalBacklogs, forward)
}

func AssignBatch(profile *interfaces.GenericField[string], institute *studentModel.Student, forward bool) {
	profile.IsNull = reflect2.IsNil(institute.Batch)
	profile.DataType = constants.TYPE_STRING
	if !profile.IsNull {
		profile.Value = strconv.Itoa(institute.Batch.StartYear) + "-" + strconv.Itoa(institute.Batch.EndYear)
	}
}

func MapSocialProfiles(profile *interfaces.SocialProfiles, socials *studentModel.SocialProfiles, forward bool) {
	AssignSocialProfile(&profile.LinkedIn, &socials.LinkedIn, forward)
	AssignSocialProfile(&profile.Github, &socials.Github, forward)
	AssignSocialProfile(&profile.CodeChef, &socials.CodeChef, forward)
	AssignSocialProfile(&profile.Codeforces, &socials.Codeforces, forward)
	AssignSocialProfile(&profile.Leetcode, &socials.LeetCode, forward)
	AssignSocialProfile(&profile.GoogleScholar, &socials.GoogleScholar, forward)
	AssignSocialProfile(&profile.MicrosoftTeams, &socials.MicrosoftTeams, forward)
	AssignSocialProfile(&profile.Kaggle, &socials.Kaggle, forward)
	AssignSocialProfile(&profile.Skype, &socials.Skype, forward)
}

func MapProfileInstitute(profile *interfaces.ProfileInstitute, institute *studentModel.Student, forward bool) {
	AssignBatch(&profile.Batch, institute, forward)
	AssignNotNilValue(&profile.RollNumber, &institute.RollNo, forward)
	AssignNotNilValue(&profile.InstituteEmail, &institute.InstituteEmail, forward)
	AssignNotNilValue(&profile.Department, &institute.Department, forward)
	AssignNilPossibleValue(&profile.EducationGap, &institute.Academics.EducationGap, forward)
	AssignNilPossibleValue(&profile.Course, &institute.Course, forward)
	AssignNilPossibleValue(&profile.Specialisation, &institute.Specialisation, forward)
	AssignNilPossibleValue(&profile.Honours, &institute.Academics.Honours, forward)
	AssignNilPossibleValue(&profile.ThesisEndDate, &institute.Academics.ThesisEndDate, forward)

	if forward {
		profile.RollNumber.IsLocked = true
		profile.Batch.IsLocked = true
		profile.InstituteEmail.IsLocked = true
		profile.Department.IsLocked = true
	}
}

func MapPastAcademics(profile *interfaces.ProfilePastAcademics, institute *studentModel.Academics, forward bool) {
	AssignPastAcademics(&profile.ClassX, &institute.XthClass, forward)
	AssignPastAcademics(&profile.ClassXII, &institute.XIIthClass, forward)
	AssignPastAcademics(&profile.Undergraduate, &institute.UnderGraduate, forward)
	AssignPastAcademics(&profile.Postgraduate, &institute.PostGraduate, forward)
	AssignPastAcademics(&profile.Diploma, &institute.Diploma, forward)
}

func MapRanks(profile *interfaces.ProfilePastAcademics, rank *studentModel.Academics, forward bool) {
	AssignRankValue(&profile.JeeRank, &rank.JEERank, forward)
	AssignRankValue(&profile.GateRank, &rank.GATERank, forward)
}

// 'forward' defines the mapping direction
// if true, then maps model (backend) to profile (frontend)
// if false, then maps profile (frontend) to model (backend)
func MapStudentToStudentProfile(profile *interfaces.StudentProfile, student *studentModel.Student, forward bool) {
	// Profile
	MapProfilePersonal(&profile.Profile.PersonalProfile, student, forward)
	MapSocialProfiles(&profile.SocialProfiles, &student.SocialProfiles, forward)
	MapProfileInstitute(&profile.Profile.InstituteProfile, student, forward)

	// Past Academics
	MapPastAcademics(&profile.PastAcademics, &student.Academics, forward)
	MapRanks(&profile.PastAcademics, &student.Academics, forward)

	// Current Academics
	MapProfileCurrentAcademics(&profile.CurrentAcademics, &student.Academics, forward)
}
