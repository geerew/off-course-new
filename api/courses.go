package api

import (
	"database/sql"
	"net/url"
	"os"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/jobs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courses struct {
	appFs             *appFs.AppFs
	courseScanner     *jobs.CourseScanner
	courseDao         *daos.CourseDao
	courseProgressDao *daos.CourseProgressDao
	assetDao          *daos.AssetDao
	attachmentDao     *daos.AttachmentDao
	tagDao            *daos.TagDao
	courseTagDao      *daos.CourseTagDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseResponse struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Path      string         `json:"path"`
	HasCard   bool           `json:"hasCard"`
	Available bool           `json:"available"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`

	// Scan status
	ScanStatus string `json:"scanStatus"`

	// Progress
	Started           bool           `json:"started"`
	StartedAt         types.DateTime `json:"startedAt"`
	Percent           int            `json:"percent"`
	CompletedAt       types.DateTime `json:"completedAt"`
	ProgressUpdatedAt types.DateTime `json:"progressUpdatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseTagResponse struct {
	ID  string `json:"id"`
	Tag string `json:"tag"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func bindCoursesApi(router fiber.Router, appFs *appFs.AppFs, db database.Database, courseScanner *jobs.CourseScanner) {
	api := courses{
		appFs:             appFs,
		courseScanner:     courseScanner,
		courseDao:         daos.NewCourseDao(db),
		courseProgressDao: daos.NewCourseProgressDao(db),
		assetDao:          daos.NewAssetDao(db),
		attachmentDao:     daos.NewAttachmentDao(db),
		tagDao:            daos.NewTagDao(db),
		courseTagDao:      daos.NewCourseTagDao(db),
	}

	subGroup := router.Group("/courses")

	// Courses
	subGroup.Get("", api.getCourses)
	subGroup.Get("/:id", api.getCourse)
	subGroup.Post("", api.createCourse)
	subGroup.Delete("/:id", api.deleteCourse)

	// Card
	subGroup.Head("/:id/card", api.getCard)
	subGroup.Get("/:id/card", api.getCard)

	// Assets
	subGroup.Get("/:id/assets", api.getAssets)
	subGroup.Get("/:id/assets/:asset", api.getAsset)

	// Attachments
	subGroup.Get("/:id/assets/:asset/attachments", api.getAssetAttachments)
	subGroup.Get("/:id/assets/:asset/attachments/:attachment", api.getAssetAttachment)

	// Tags
	subGroup.Get("/:id/tags", api.getTags)
	subGroup.Post("/:id/tags", api.createTag)
	subGroup.Delete("/:id/tags/:tagId", api.deleteTag)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getCourses(c *fiber.Ctx) error {
	orderBy := c.Query("orderBy", "created_at desc")
	progress := c.Query("progress", "")
	tags := c.Query("tags", "")
	titles := c.Query("titles", "")

	dbParams := &database.DatabaseParams{
		OrderBy:    strings.Split(orderBy, ","),
		Pagination: pagination.NewFromApi(c),
	}

	whereClause := squirrel.And{}

	// Filter on progress (one of "not started", "started", "completed")
	if progress != "" {
		unescapedProgress, err := url.QueryUnescape(progress)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid progress parameter")
		}

		unescapedProgress = strings.ToLower(unescapedProgress)

		if unescapedProgress == "started" {
			// Started but not completed
			whereClause = append(whereClause, squirrel.And{
				squirrel.Eq{api.courseProgressDao.Table() + ".started": true},
				squirrel.NotEq{api.courseProgressDao.Table() + ".percent": 100},
			})
		} else if unescapedProgress == "not started" {
			// Default to not started
			whereClause = append(whereClause, squirrel.NotEq{api.courseProgressDao.Table() + ".started": true})
		} else if unescapedProgress == "completed" {
			// Completed
			whereClause = append(whereClause, squirrel.Eq{api.courseProgressDao.Table() + ".percent": 100})
		} else if unescapedProgress == "not completed" {
			// Not completed
			whereClause = append(whereClause, squirrel.NotEq{api.courseProgressDao.Table() + ".percent": 100})
		}
	}

	// Filter based on tags
	if tags != "" {
		unescapedTags, err := url.QueryUnescape(tags)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid tags parameter")
		}

		tagList := strings.Split(unescapedTags, ",")

		courseIds, err := api.courseTagDao.ListCourseIdsByTags(tagList, nil)
		if err != nil {
			log.Err(err).Msg("error looking up course ids by tags")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "error looking up courses by tags - " + err.Error(),
			})
		}

		if len(courseIds) == 0 {
			pResult, err := dbParams.Pagination.BuildResult(courseResponseHelper(nil))
			if err != nil {
				log.Err(err).Msg("error building pagination result")
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "error building pagination result - " + err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(pResult)
		} else {
			whereClause = append(whereClause, squirrel.Eq{api.courseDao.Table() + ".id": courseIds})
		}
	}

	// Filter based on titles
	if titles != "" {
		unescapedTitles, err := url.QueryUnescape(titles)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid titles parameter")
		}

		titleList := strings.Split(unescapedTitles, ",")

		orClause := squirrel.Or{}
		for _, title := range titleList {
			orClause = append(orClause, squirrel.Like{api.courseDao.Table() + ".title": "%" + title + "%"})
		}

		whereClause = append(whereClause, orClause)
	}

	dbParams.Where = whereClause

	courses, err := api.courseDao.List(dbParams, nil)

	if err != nil {
		log.Err(err).Msg("error looking up courses")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up courses - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(courseResponseHelper(courses))
	if err != nil {
		log.Err(err).Msg("error building pagination result")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error building pagination result - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	course, err := api.courseDao.Get(id, nil, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(courseResponseHelper([]*models.Course{course})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) createCourse(c *fiber.Ctx) error {
	course := new(models.Course)

	if err := c.BodyParser(course); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "error parsing data - " + err.Error(),
		})
	}

	// Ensure there is a title and path
	if course.Title == "" || course.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "a title and path are required",
		})
	}

	// Empty stuff that should not be set
	course.ID = ""

	// Validate the path
	if exists, err := afero.DirExists(api.appFs.Fs, course.Path); err != nil || !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid course path",
		})
	}

	// Set the course to available
	course.Available = true

	if err := api.courseDao.Create(course); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "a course with this path already exists - " + err.Error(),
			})
		}

		log.Err(err).Msg("error creating course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error creating course - " + err.Error(),
		})
	}

	// Start a scan job
	if scan, err := api.courseScanner.Add(course.ID); err != nil {
		log.Err(err).Msg("error creating scan job")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error creating scan job - " + err.Error(),
		})
	} else {
		course.ScanStatus = scan.Status.String()
	}

	return c.Status(fiber.StatusCreated).JSON(courseResponseHelper([]*models.Course{course})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) deleteCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	err := api.courseDao.Delete(&database.DatabaseParams{Where: squirrel.Eq{"id": id}}, nil)
	if err != nil {
		log.Err(err).Msg("error deleting course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error deleting course - " + err.Error(),
		})

	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getCard(c *fiber.Ctx) error {
	id := c.Params("id")

	course, err := api.courseDao.Get(id, nil, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Course not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	if course.CardPath == "" {
		return c.Status(fiber.StatusNotFound).SendString("Course has no card")
	}

	_, err = api.appFs.Fs.Stat(course.CardPath)
	if os.IsNotExist(err) {
		log.Err(err).Str("card", course.CardPath).Msg("card not found on disk")
		return c.Status(fiber.StatusNotFound).SendString("Course card not found")
	}

	// The fiber function sendFile(...) does not support using a custom FS. Therefore, use
	// SendFile() from the filesystem middleware.
	return filesystem.SendFile(c, afero.NewHttpFs(api.appFs.Fs), course.CardPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getAssets(c *fiber.Ctx) error {
	id := c.Params("id")
	orderBy := c.Query("orderBy", "chapter asc,prefix asc")
	expand := c.QueryBool("expand", false)

	// Get the course
	_, err := api.courseDao.Get(id, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	dbParams := &database.DatabaseParams{
		OrderBy:    strings.Split(orderBy, ","),
		Where:      squirrel.Eq{api.assetDao.Table() + ".course_id": id},
		Pagination: pagination.NewFromApi(c),
	}

	if expand {
		dbParams.IncludeRelations = []string{api.attachmentDao.Table()}
	}

	assets, err := api.assetDao.List(dbParams, nil)
	if err != nil {
		log.Err(err).Msg("error looking up assets")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up assets - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(assetResponseHelper(assets))
	if err != nil {
		log.Err(err).Msg("error building pagination result")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error building pagination result - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getAsset(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	expand := c.QueryBool("expand", false)

	_, err := api.courseDao.Get(id, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	// TODO: support attachments orderby
	dbParams := &database.DatabaseParams{}
	if expand {
		dbParams.IncludeRelations = []string{api.attachmentDao.Table()}
	}

	asset, err := api.assetDao.Get(assetId, dbParams, nil)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up asset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up asset - " + err.Error(),
		})
	}

	if asset.CourseID != id {
		return c.Status(fiber.StatusBadRequest).SendString("Asset does not belong to course")
	}

	return c.Status(fiber.StatusOK).JSON(assetResponseHelper([]*models.Asset{asset})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getAssetAttachments(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	orderBy := c.Query("orderBy", "title asc")

	// Get the course
	_, err := api.courseDao.Get(id, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	// Get the asset
	asset, err := api.assetDao.Get(assetId, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up asset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up asset - " + err.Error(),
		})
	}

	if asset.CourseID != id {
		return c.Status(fiber.StatusBadRequest).SendString("Asset does not belong to course")
	}

	dbParams := &database.DatabaseParams{
		OrderBy:    strings.Split(orderBy, ","),
		Where:      squirrel.Eq{api.attachmentDao.Table() + ".asset_id": assetId},
		Pagination: pagination.NewFromApi(c),
	}

	attachments, err := api.attachmentDao.List(dbParams, nil)
	if err != nil {
		log.Err(err).Msg("error looking up attachments")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up attachments - " + err.Error(),
		})
	}

	pResult, err := dbParams.Pagination.BuildResult(attachmentResponseHelper(attachments))
	if err != nil {
		log.Err(err).Msg("error building pagination result")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error building pagination result - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getAssetAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	attachmentId := c.Params("attachment")

	// Get the course
	_, err := api.courseDao.Get(id, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	// Get the asset
	asset, err := api.assetDao.Get(assetId, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up asset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up asset - " + err.Error(),
		})
	}

	if asset.CourseID != id {
		return c.Status(fiber.StatusBadRequest).SendString("Asset does not belong to course")
	}

	// Get the attachment
	attachment, err := api.attachmentDao.Get(attachmentId, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up attachment")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up attachment - " + err.Error(),
		})
	}

	if attachment.AssetID != assetId {
		return c.Status(fiber.StatusBadRequest).SendString("Attachment does not belong to asset")
	}

	return c.Status(fiber.StatusOK).JSON(attachmentResponseHelper([]*models.Attachment{attachment})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) getTags(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get the course
	_, err := api.courseDao.Get(id, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		log.Err(err).Msg("error looking up course")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course - " + err.Error(),
		})
	}

	dbParams := &database.DatabaseParams{
		OrderBy: []string{api.tagDao.Table() + ".tag asc"},
		Where:   squirrel.Eq{api.courseTagDao.Table() + ".course_id": id},
	}

	tags, err := api.courseTagDao.List(dbParams, nil)
	if err != nil {
		log.Err(err).Msg("error looking up course tags")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error looking up course tags - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(courseTagResponseHelper(tags))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) createTag(c *fiber.Ctx) error {
	courseId := c.Params("id")
	courseTag := new(models.CourseTag)

	if err := c.BodyParser(courseTag); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "error parsing data - " + err.Error(),
		})
	}

	// Empty stuff that should not be set
	courseTag.ID = ""
	courseTag.TagId = ""

	// Ensure there is a tag
	if courseTag.Tag == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "a tag is required",
		})
	}

	// Set the course ID
	courseTag.CourseId = courseId

	if err := api.courseTagDao.Create(courseTag, nil); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "a tag for this course already exists - " + err.Error(),
			})

		}

		log.Err(err).Msg("error creating course tag")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error creating course tag - " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(courseTagResponseHelper([]*models.CourseTag{courseTag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api *courses) deleteTag(c *fiber.Ctx) error {
	courseId := c.Params("id")
	tagId := c.Params("tagId")

	err := api.courseTagDao.Delete(&database.DatabaseParams{Where: squirrel.And{squirrel.Eq{"course_id": courseId}, squirrel.Eq{"id": tagId}}}, nil)
	if err != nil {
		log.Err(err).Msg("error deleting course tag")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error deleting course tag - " + err.Error(),
		})

	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPER
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func courseResponseHelper(courses []*models.Course) []*courseResponse {
	responses := []*courseResponse{}
	for _, course := range courses {
		c := &courseResponse{
			ID:        course.ID,
			Title:     course.Title,
			Path:      course.Path,
			HasCard:   course.CardPath != "",
			Available: course.Available,
			CreatedAt: course.CreatedAt,
			UpdatedAt: course.UpdatedAt,

			// Scan status
			ScanStatus: course.ScanStatus,

			// Progress
			Started:           course.Started,
			StartedAt:         course.StartedAt,
			Percent:           course.Percent,
			CompletedAt:       course.CompletedAt,
			ProgressUpdatedAt: course.ProgressUpdatedAt,
		}

		responses = append(responses, c)
	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func courseTagResponseHelper(courseTags []*models.CourseTag) []*courseTagResponse {
	responses := []*courseTagResponse{}
	for _, tag := range courseTags {
		responses = append(responses, &courseTagResponse{
			ID:  tag.ID,
			Tag: tag.Tag,
		})
	}

	return responses
}
