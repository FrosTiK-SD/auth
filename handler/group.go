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
)

func (h *Handler) GetAllGroups(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)
	groups, err := controller.GetAllGroups(h.MongikClient, noCache)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": err,
			"data":  nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  groups,
		"error": nil,
	})
}

func (h *Handler) BatchCreateGroup(ctx *gin.Context) {
	batchCreateGroupRequest := interfaces.BatchCreateGroupRequest{}

	if errBinding := ctx.BindJSON(&batchCreateGroupRequest); errBinding != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_INCORRENT_BODY,
			"message": errBinding,
		})
		return
	}
	insertResult, err := controller.BatchCreateGroup(h.MongikClient, batchCreateGroupRequest.Groups)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_MONGO_ERROR,
			"message": err,
		})
		return
	}

	admin, exists := ctx.Get(constants.SESSION)
	if exists {
		adminStudent := admin.(*model.StudentPopulated)
		h.LogActivityDirect(adminStudent.Id, "CREATE", fmt.Sprintf("Batch created %d groups", len(batchCreateGroupRequest.Groups)))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": insertResult,
	})
}

func (h *Handler) BatchEditGroup(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)

	assignRequests := []interfaces.AssignRequest{}

	if errBinding := ctx.BindJSON(&assignRequests); errBinding != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_INCORRENT_BODY,
			"message": errBinding,
		})
		return
	}

	addResult, removeResult, errors := controller.BatchEditGroup(h.MongikClient, assignRequests, noCache)

	if len(*errors) != 0 {
		ctx.JSON(http.StatusPartialContent, gin.H{
			"data": gin.H{
				"addList":    addResult,
				"removeList": removeResult,
			},
			"error": errors,
		})
	} else {
		admin, exists := ctx.Get(constants.SESSION)
		if exists {
			adminStudent := admin.(*model.StudentPopulated)
			h.LogActivityDirect(adminStudent.Id, "EDIT", "Batch edited groups (assigned/unassigned roles)")
		}
		ctx.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"addList":    addResult,
				"removeList": removeResult,
			},
			"error": nil,
		})
	}

}

func (h *Handler) BatchDeleteGroup(ctx *gin.Context) {
	batchDeleteGroupRequest := interfaces.BatchDeleteGroupRequest{}

	if errBinding := ctx.BindJSON(&batchDeleteGroupRequest); errBinding != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_INCORRENT_BODY,
			"message": errBinding,
		})
		return
	}

	groupResult, studentResult, err := controller.BatchDeleteGroup(h.MongikClient, &batchDeleteGroupRequest.Groups)

	if *err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"group":    groupResult,
				"students": studentResult,
			},
			"error":   constants.ERROR_MONGO_ERROR,
			"message": err,
		})
		return
	}

	admin, exists := ctx.Get(constants.SESSION)
	if exists {
		adminStudent := admin.(*model.StudentPopulated)
		h.LogActivityDirect(adminStudent.Id, "DELETE", fmt.Sprintf("Batch deleted %d groups", len(batchDeleteGroupRequest.Groups)))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"group":    groupResult,
			"students": studentResult,
		},
		"error": nil,
	})
}

func (h *Handler) BatchAssignGroup(ctx *gin.Context) {
	batchAssignGroupRequest := []interfaces.BatchAssignGroupRequest{}

	if errBinding := ctx.BindJSON(&batchAssignGroupRequest); errBinding != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_INCORRENT_BODY,
			"message": errBinding,
		})
		return
	}
	addList, removeList, errors := controller.BatchAssignGroup(h.MongikClient, batchAssignGroupRequest)

	if len(errors) != 0 {
		ctx.AbortWithStatusJSON(http.StatusPartialContent, gin.H{
			"data": gin.H{
				"addList":    addList,
				"removeList": removeList,
			},
			"error": errors,
		})
		return
	}

	admin, exists := ctx.Get(constants.SESSION)
	if exists {
		adminStudent := admin.(*model.StudentPopulated)
		h.LogActivityDirect(adminStudent.Id, "EDIT", fmt.Sprintf("Batch assigned/unassigned groups for %d students", len(batchAssignGroupRequest)))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"addList":    addList,
			"removeList": removeList,
		},
		"error": nil,
	})
}
