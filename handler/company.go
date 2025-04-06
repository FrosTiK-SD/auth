package handler

import (
	"net/http"
	"strconv"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/controller"
	"github.com/FrosTiK-SD/auth/util"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAllCompanies(ctx *gin.Context) {
	noCache := util.GetNoCache(ctx)
	currentPage := 0

	currentPage, err := strconv.Atoi(ctx.Request.URL.Query().Get("page"))
	companiesPerPage, err := strconv.Atoi(ctx.Request.URL.Query().Get("limit"))

	companies, totalCompanies, err := controller.GetAllCompanies(h.MongikClient, noCache, currentPage, companiesPerPage)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   constants.ERROR_MONGO_ERROR,
			"message": err,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":           companies,
		"totalCompanies": totalCompanies,
	})
}
