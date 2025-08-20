package handler

import (
	"errors"
	"fmt"

	"github.com/FrosTiK-SD/auth/constants"
	"github.com/FrosTiK-SD/auth/controller"
	"github.com/FrosTiK-SD/auth/interfaces"
	"github.com/FrosTiK-SD/auth/model"
	"github.com/FrosTiK-SD/auth/util"
	"github.com/gofiber/fiber/v2"
)

// For Gin based middlewares
func (h *Handler) FiberVerifyStudent(ctx *fiber.Ctx) error {
	idToken := ctx.Get("token", "")
	noCache := false
	if ctx.Get("cache-control") == constants.NO_CACHE {
		noCache = true
	}

	email, _, err := controller.VerifyToken(h.MongikClient.CacheClient, idToken, h.JwkSet, noCache)

	if err != nil {
		return errors.New(*err)
	}
	student, err := controller.GetUserByEmail(h.MongikClient, email, &constants.ROLE_STUDENT, noCache)
	if err != nil {
		return errors.New(*err)
	}

	ctx.Locals(constants.SESSION, student)

	return ctx.Next()
}

// To be used only after FiberVerifyStudent
func (h *Handler) FiberGetRoleCheckHandlerForStudent(roles ...string) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		student, ok := ctx.Locals(constants.SESSION).(*model.StudentPopulated)

		if !ok {
			return ctx.Status(fiber.StatusForbidden).JSON(
				map[string]any{
					"message": constants.ERROR_ROLE_CHECK_FAILED,
					"error":   "Student does not exist",
				},
			)
		}

		for _, role := range roles {
			if !util.CheckRoleExists(&student.GroupDetails, role) {
				return ctx.Status(fiber.StatusForbidden).JSON(
					map[string]any{
						"message": constants.ERROR_ROLE_CHECK_FAILED,
						"error":   fmt.Sprintf("Student Does not have Role '%s'", role),
					},
				)
			}
		}

		return ctx.Next()
	}
}

func (h *RoleCheckerHandler) FiberVerifyRole(ctx *fiber.Ctx) error {
	entity := ctx.Locals(constants.SESSION)
	var entityGroups *interfaces.Groups
	entityBytes, err := json.Marshal(entity)
	if err != nil {
		return err
	}
	err = json.Unmarshal(entityBytes, &entityGroups)
	if err != nil {
		return err
	}
	if !util.CheckRoleExists(&entityGroups.Groups, h.Role) {
		return errors.New(constants.ERROR_ROLE_CHECK_FAILED)
	}

	ctx.Next()
	return nil
}
