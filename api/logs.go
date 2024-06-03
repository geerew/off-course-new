package api

import (
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type logs struct {
	logger *slog.Logger
	logDao *daos.LogDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type logResponse struct {
	ID        string         `json:"id"`
	Level     int            `json:"level"`
	Message   string         `json:"message"`
	Data      types.JsonMap  `json:"data"`
	CreatedAt types.DateTime `json:"createdAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *logs) getLogTypes(c *fiber.Ctx) error {
	types := types.AllLogTypes()
	return c.Status(fiber.StatusOK).JSON(types)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *logs) getLogs(c *fiber.Ctx) error {
	levels := c.Query("levels", "")
	types := c.Query("types", "")
	messages := c.Query("messages", "")

	dbParams := &database.DatabaseParams{
		Pagination: pagination.NewFromApi(c),
	}

	whereClause := squirrel.And{}

	// Log levels
	if levels != "" {
		filtered, err := filter(levels)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid levels parameter", err)
		}

		whereClause = append(whereClause, squirrel.Eq{api.logDao.Table() + ".level": filtered})
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
			orClause = append(orClause, squirrel.Like{api.logDao.Table() + ".message": "%" + message + "%"})
		}

		whereClause = append(whereClause, orClause)
	}

	dbParams.Where = whereClause

	logs, err := api.logDao.List(dbParams, nil)

	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up logs", err)
	}

	pResult, err := dbParams.Pagination.BuildResult(logsResponseHelper(logs))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func logsResponseHelper(logs []*models.Log) []*logResponse {
	responses := []*logResponse{}

	for _, log := range logs {
		responses = append(responses, &logResponse{
			ID:        log.ID,
			Level:     log.Level,
			Message:   log.Message,
			Data:      log.Data,
			CreatedAt: log.CreatedAt,
		})
	}

	return responses
}
