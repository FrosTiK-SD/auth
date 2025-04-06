package interfaces

import (
	constantModel "github.com/FrosTiK-SD/models/constant"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TYPE_SOCIAL = string

type GenericField[T bool | string | int | float64 | primitive.DateTime | constantModel.Branch | constantModel.Course | constantModel.Gender] struct {
	DataType    string `json:"dataType"`
	DataChoices *[]T   `json:"dataChoices,omitempty"`
	IsVerified  *bool  `json:"isVerified,omitempty"`
	Value       T      `json:"value"`
	IsLocked    bool   `json:"isLocked"`
	IsRequired  bool   `json:"isRequired"`
	IsNull      bool   `json:"isNull"`
	IsHidden    bool   `json:"isHidden"`
}

type ProfilePersonal struct {
	FirstName        GenericField[string]               `json:"firstName,omitempty"`
	MiddleName       GenericField[string]               `json:"middleName,omitempty"`
	LastName         GenericField[string]               `json:"lastName,omitempty"`
	Gender           GenericField[constantModel.Gender] `json:"gender,omitempty"`
	DOB              GenericField[primitive.DateTime]   `json:"dob,omitempty"`
	PermanentAddress GenericField[string]               `json:"permanentAddress,omitempty"`
	PresentAddress   GenericField[string]               `json:"presentAddress,omitempty"`
	PersonalEmail    GenericField[string]               `json:"personalEmail,omitempty"`
	Mobile           GenericField[string]               `json:"mobile,omitempty"`
	Category         GenericField[string]               `json:"category,omitempty"`
	IsPWD            GenericField[bool]                 `json:"isPWD,omitempty"`
	IsEWS            GenericField[bool]                 `json:"isEWS,omitempty"`
	FatherName       GenericField[string]               `json:"fatherName,omitempty"`
	FatherOccupation GenericField[string]               `json:"fatherOccupation,omitempty"`
	MotherName       GenericField[string]               `json:"motherName,omitempty"`
	MotherOccupation GenericField[string]               `json:"motherOccupation,omitempty"`
	MotherTongue     GenericField[string]               `json:"motherTongue,omitempty"`
}

type SocialProfile struct {
	URL      GenericField[string] `json:"url,omitempty"`
	Username GenericField[string] `json:"username,omitempty"`
}

type SocialProfiles struct {
	LinkedIn       SocialProfile `json:"linkedIn,omitempty"`
	Github         SocialProfile `json:"github,omitempty"`
	Kaggle         SocialProfile `json:"kaggle,omitempty"`
	MicrosoftTeams SocialProfile `json:"microsoftTeams,omitempty"`
	Skype          SocialProfile `json:"skype,omitempty"`
	GoogleScholar  SocialProfile `json:"googleScholar,omitempty"`
	Codeforces     SocialProfile `json:"codeforces,omitempty"`
	CodeChef       SocialProfile `json:"codechef,omitempty"`
	Leetcode       SocialProfile `json:"leetcode,omitempty"`
}

type GenericRank struct {
	Rank     GenericField[int]    `json:"rank,omitempty"`
	Category GenericField[string] `json:"category,omitempty"`
	IsEWS    GenericField[bool]   `json:"isEWS,omitempty"`
	IsPWD    GenericField[bool]   `json:"isPWD,omitempty"`
}

type ProfileInstitute struct {
	RollNumber     GenericField[int]                  `json:"rollNo,omitempty"`
	InstituteEmail GenericField[string]               `json:"email,omitempty"`
	Batch          GenericField[string]               `json:"batch,omitempty"`
	Department     GenericField[string]               `json:"department,omitempty"`
	Course         GenericField[constantModel.Course] `json:"course,omitempty"`
	Specialisation GenericField[string]               `json:"specialisation,omitempty"`
	Honours        GenericField[string]               `json:"honours,omitempty"`
	ThesisEndDate  GenericField[primitive.DateTime]   `json:"thesisEndDate,omitempty"`
	EducationGap   GenericField[int]                  `json:"educationGap,omitempty"`
}

type ProfileDetails struct {
	PersonalProfile  ProfilePersonal  `json:"personal_profile,omitempty"`
	InstituteProfile ProfileInstitute `json:"institute_profile,omitempty"`
}

type ProfilePastEducation struct {
	Certification GenericField[string]  `json:"certification,omitempty"`
	Institute     GenericField[string]  `json:"institute,omitempty"`
	Year          GenericField[int]     `json:"year,omitempty"`
	Score         GenericField[float64] `json:"score,omitempty"`
}

type ProfilePastAcademics struct {
	ClassX        ProfilePastEducation `json:"class_x,omitempty"`
	ClassXII      ProfilePastEducation `json:"class_xii,omitempty"`
	Undergraduate ProfilePastEducation `json:"undergraduate,omitempty"`
	Postgraduate  ProfilePastEducation `json:"postgraduate,omitempty"`
	JeeRank       GenericRank          `json:"jeeRank,omitempty"`
	GateRank      GenericRank          `json:"gateRank,omitempty"`
	Diploma		  ProfilePastEducation `json:"diploma,omitempty"`
}

type ProfileCurrentAcademicsMisc struct {
	CurrentCGPA    GenericField[float64] `json:"current_cgpa,omitempty"`
	ActiveBacklogs GenericField[int]     `json:"active_backlogs,omitempty"`
	TotalBacklogs  GenericField[int]     `json:"total_backlogs,omitempty"`
}

type ProfileSemesterSPI struct {
	One   GenericField[float64] `json:"one,omitempty"`
	Two   GenericField[float64] `json:"two,omitempty"`
	Three GenericField[float64] `json:"three,omitempty"`
	Four  GenericField[float64] `json:"four,omitempty"`
	Five  GenericField[float64] `json:"five,omitempty"`
	Six   GenericField[float64] `json:"six,omitempty"`
	Seven GenericField[float64] `json:"seven,omitempty"`
	Eight GenericField[float64] `json:"eight,omitempty"`
	Nine  GenericField[float64] `json:"nine,omitempty"`
	Ten   GenericField[float64] `json:"ten,omitempty"`
}

type ProfileSummerTermSPI struct {
	One   GenericField[float64] `json:"one,omitempty"`
	Two   GenericField[float64] `json:"two,omitempty"`
	Three GenericField[float64] `json:"three,omitempty"`
	Four  GenericField[float64] `json:"four,omitempty"`
	Five  GenericField[float64] `json:"five,omitempty"`
}

type ProfileCurrentAcademics struct {
	SemesterSPI   ProfileSemesterSPI          `json:"semester_spi,omitempty"`
	SummerTermSPI ProfileSummerTermSPI        `json:"summer_term_spi,omitempty"`
	Misc          ProfileCurrentAcademicsMisc `json:"misc,omitempty"`
}

type StudentProfile struct {
	Profile          ProfileDetails          `json:"profile,omitempty"`
	SocialProfiles   SocialProfiles          `json:"social_profiles,omitempty"`
	PastAcademics    ProfilePastAcademics    `json:"past_academics,omitempty"`
	CurrentAcademics ProfileCurrentAcademics `json:"current_academics,omitempty"`
}
