package api

import (
	"log/slog"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type logsAPI struct {
	logger *slog.Logger
	dao    *dao.DAO
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initLogRoutes initializes the log routes
func (r *Router) initLogRoutes() {
	logsAPI := logsAPI{
		logger: r.config.Logger,
		dao:    r.logDao,
	}

	logGroup := r.api.Group("/logs")
	logGroup.Get("/", logsAPI.getLogs)
	logGroup.Get("/types", logsAPI.getLogTypes)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *logsAPI) getLogs(c *fiber.Ctx) error {
	orderBy := c.Query("orderBy", models.LOG_TABLE+".created_at desc")
	levels := c.Query("levels", "")
	types := c.Query("types", "")
	messages := c.Query("messages", "")

	options := &database.Options{
		OrderBy:    strings.Split(orderBy, ","),
		Pagination: pagination.NewFromApi(c),
	}

	whereClause := squirrel.And{}

	// Log levels
	if levels != "" {
		filtered, err := filter(levels)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid levels parameter", err)
		}

		whereClause = append(whereClause, squirrel.Eq{models.LOG_TABLE + ".level": filtered})
	}

	// Log types
	if types != "" {
		filtered, err := filter(types)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid types parameter", err)
		}

		whereClause = append(whereClause, squirrel.Eq{"JSON_EXTRACT(data, '$.type')": filtered})
	}

	// Log message
	if messages != "" {
		filtered, err := filter(messages)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid messages parameter", err)
		}

		orClause := squirrel.Or{}
		for _, message := range filtered {
			orClause = append(orClause, squirrel.Like{models.LOG_TABLE + ".message": "%" + message + "%"})
		}

		whereClause = append(whereClause, orClause)
	}

	options.Where = whereClause

	logs := []*models.Log{}
	err := api.dao.List(c.UserContext(), &logs, options)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up logs", err)
	}

	pResult, err := options.Pagination.BuildResult(logsResponseHelper(logs))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *logsAPI) getLogTypes(c *fiber.Ctx) error {
	types := types.AllLogTypes()
	return c.Status(fiber.StatusOK).JSON(types)
}
