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

func (api *logs) getLogs(c *fiber.Ctx) error {
	minLevel := c.QueryInt("level", -4)

	dbParams := &database.DatabaseParams{
		Where:      squirrel.GtOrEq{api.logDao.Table() + ".level": minLevel},
		Pagination: pagination.NewFromApi(c),
	}

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
