package handler

import (
	"errors"
	"net/http"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/controller"
	"github.com/FrosTiK-SD/auth/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) HandlerVerifyStudentIdToken(ctx *gin.Context) {
	idToken := ctx.GetHeader("token")
	noCache := false
	if ctx.GetHeader("cache-control") == constants.NO_CACHE {
		noCache = true
	}

	email, exp, err := controller.VerifyToken(h.MongikClient.CacheClient, idToken, h.JwkSet, noCache)

	if err != nil {
		if h.Config.Mode == MIDDLEWARE {
			h.Session.Error = errors.New(*err)
		}
		ctx.JSON(200, gin.H{
			"student": nil,
			"expire":  exp,
			"error":   err,
		})
		return
	}

	student, err := controller.GetUserByEmail(h.MongikClient, email, &constants.ROLE_STUDENT, noCache)
	if err != nil {
		if h.Config.Mode == MIDDLEWARE {
			h.Session.Error = errors.New(*err)
		}
		ctx.JSON(200, gin.H{
			"data":   nil,
			"error":  err,
			"expire": exp,
		})
		return
	}

	impersonateId := ctx.GetHeader("x-impersonate-student-id")
	if impersonateId == "" {
		impersonateId = ctx.GetHeader("X-Impersonate-Student-Id")
	}

	if impersonateId != "" {
		if util.CheckRoleExists(&student.GroupDetails, constants.ROLE_OPPORTUNITIES_WRITE) {
			targetObjId, parseErr := primitive.ObjectIDFromHex(impersonateId)
			if parseErr == nil {
				targetStudent, targetErr := controller.GetStudentById(h.MongikClient, targetObjId, noCache)
				if targetErr == nil && targetStudent != nil {
					student = targetStudent
				}
			}
		} else {
			errMsg := "Unauthorized impersonation attempt"
			if h.Config.Mode == MIDDLEWARE {
				h.Session.Error = errors.New(errMsg)
			}
			ctx.JSON(http.StatusForbidden, gin.H{
				"data":   nil,
				"error":  &errMsg,
				"expire": exp,
			})
			return
		}
	}

	if h.Config.Mode == MIDDLEWARE {
		h.Session.Student = student
	} else {
		ctx.JSON(200, gin.H{
			"data":   student,
			"error":  nil,
			"expire": exp,
		})
	}
}

func (h *Handler) HandlerVerifyRecruiterIdToken(ctx *gin.Context) {
	idToken := ctx.GetHeader("token")
	noCache := false
	if ctx.GetHeader("cache-control") == constants.NO_CACHE {
		noCache = true
	}
	email, _, err := controller.VerifyToken(h.MongikClient.CacheClient, idToken, h.JwkSet, noCache)

	if email != nil && *email != "" {
		recruiter, recErr := controller.GetRecruiterByEmail(h.MongikClient, email, &constants.ROLE_RECRUITER, noCache)
		if recErr != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"data": nil,
				"error": recErr,
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"data": recruiter,
		})
		return
	}

	// If email is nil or empty, return error/status like TypeScript
	status := 500
	if err != nil && len(*err) >= 4 && (*err)[:4] == "auth" {
		status = 401
	}
	ctx.JSON(status, gin.H{
		"error":  err,
		"status": status,
	})
}

func (h *Handler) InvalidateCache(ctx *gin.Context) {
	h.MongikClient.CacheClient.Delete("GCP_JWKS")
	ctx.JSON(200, gin.H{
		"message": "Successfully invalidated cache",
	})
}
