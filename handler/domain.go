package handler

import (
	"fmt"
	"net/http"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/controller"
	"github.com/FrosTiK-SD/auth/interfaces"
	"github.com/FrosTiK-SD/auth/model"
	"github.com/FrosTiK-SD/auth/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) GetAllDomains(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)

	domains, err := controller.GetAllDomains(h.MongikClient, noCache)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_MONGO_ERROR,
			"message": err,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": domains,
	})
}

func (h *Handler) GetDomainById(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)
	_id, err := primitive.ObjectIDFromHex(ctx.GetHeader("id"))

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_INCORRENT_BODY,
			"message": "Invalid Id",
		})
		return
	}

	domain, err := controller.GetDomainById(h.MongikClient, _id, noCache)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_MONGO_ERROR,
			"message": err,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": domain,
	})
}

func (h *Handler) BatchCreateDomain(ctx *gin.Context) {

	batchCreateDomainRequest := interfaces.BatchCreateDomainRequest{}

	if errBinding := ctx.BindJSON(&batchCreateDomainRequest); errBinding != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_INCORRENT_BODY,
			"message": errBinding,
		})
		return
	}

	newDomains, usersList, errors := controller.BatchCreateDomain(h.MongikClient, batchCreateDomainRequest.Domains)

	if len(errors) != 0 {
		ctx.AbortWithStatusJSON(http.StatusPartialContent, gin.H{
			"data": gin.H{
				"newDomains": newDomains,
				"usersList":  usersList,
			},
			"error": errors,
		})
		return
	}
	admin, exists := ctx.Get(constants.SESSION)
	if exists {
		adminStudent := admin.(*model.StudentPopulated)
		h.LogActivityDirect(adminStudent.Id, "CREATE", fmt.Sprintf("Batch created %d domains", len(batchCreateDomainRequest.Domains)))
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"newDomains": newDomains,
			"usersList":  usersList,
		},
		"error": nil,
	})

}

func (h *Handler) EditDomainById(ctx *gin.Context) {

	domainId, err := primitive.ObjectIDFromHex(ctx.GetHeader("id"))

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_INCORRENT_BODY,
			"message": "Invalid Domain ID",
		})
		return
	}

	updateDomainRequest := interfaces.UpdateDomainRequest{}

	if errBinding := ctx.BindJSON(&updateDomainRequest); errBinding != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_INCORRENT_BODY,
			"message": errBinding,
		})
		return
	}

	oldDomain, oldStudentsResult, newStudentsResult, err := controller.UpdateDomainById(h.MongikClient, domainId, &updateDomainRequest.Domain)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusPartialContent, gin.H{
			"data": gin.H{
				"oldDomain":        oldDomain,
				"usersListOld":     oldStudentsResult,
				"usersListUpdated": newStudentsResult,
			},
			"error": err,
		})
		return
	}

	admin, exists := ctx.Get(constants.SESSION)
	if exists {
		adminStudent := admin.(*model.StudentPopulated)
		h.LogActivityDirect(adminStudent.Id, "EDIT", fmt.Sprintf("Edited domain (ID: %s)", domainId.Hex()))
	}

	ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
		"data": gin.H{
			"oldDomain":        oldDomain,
			"usersListOld":     oldStudentsResult,
			"usersListUpdated": newStudentsResult,
		},
		"error": nil,
	})
}

func (h *Handler) DeleteDomainById(ctx *gin.Context) {

	noCache := util.GetNoCache(ctx)
	domainId, err := primitive.ObjectIDFromHex(ctx.GetHeader("id"))

	if err != nil {

		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_INCORRENT_BODY,
			"message": "Invalid Domain Id",
		})
		return
	}

	deleteResult, studentResult, err := controller.DeleteDomainById(h.MongikClient, domainId, noCache)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"deleteResult":  deleteResult,
				"studentResult": studentResult,
			},
			"error":   constants.ERROR_MONGO_ERROR,
			"message": err,
		})
		return
	}

	admin, exists := ctx.Get(constants.SESSION)
	if exists {
		adminStudent := admin.(*model.StudentPopulated)
		h.LogActivityDirect(adminStudent.Id, "DELETE", fmt.Sprintf("Deleted domain (ID: %s)", domainId.Hex()))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"deleteResult":  deleteResult,
			"studentResult": studentResult,
		},
		"error": nil,
	})

}
