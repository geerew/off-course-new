package api

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/coursescan"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type coursesAPI struct {
	logger     *slog.Logger
	appFs      *appFs.AppFs
	courseScan *coursescan.CourseScan
	dao        *dao.DAO
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// initCourseRoutes initializes the course routes
func (r *Router) initCourseRoutes() {
	coursesAPI := coursesAPI{
		logger:     r.config.Logger,
		appFs:      r.config.AppFs,
		courseScan: r.config.CourseScan,
		dao:        r.dao,
	}

	courseGroup := r.api.Group("/courses")

	// Course
	courseGroup.Get("", coursesAPI.getCourses)
	courseGroup.Get("/:id", coursesAPI.getCourse)
	courseGroup.Post("", coursesAPI.createCourse)
	courseGroup.Delete("/:id", coursesAPI.deleteCourse)

	// Course card
	courseGroup.Head("/:id/card", coursesAPI.getCard)
	courseGroup.Get("/:id/card", coursesAPI.getCard)

	// Course asset
	courseGroup.Get("/:id/assets", coursesAPI.getAssets)
	courseGroup.Get("/:id/assets/:asset", coursesAPI.getAsset)
	courseGroup.Get("/:id/assets/:asset/serve", coursesAPI.serveAsset)
	courseGroup.Put("/:id/assets/:asset/progress", coursesAPI.updateAssetProgress)

	// Course asset attachments
	courseGroup.Get("/:id/assets/:asset/attachments", coursesAPI.getAttachments)
	courseGroup.Get("/:id/assets/:asset/attachments/:attachment", coursesAPI.getAttachment)
	courseGroup.Get("/:id/assets/:asset/attachments/:attachment/serve", coursesAPI.serveAttachment)

	// Course tags
	courseGroup.Get("/:id/tags", coursesAPI.getTags)
	courseGroup.Post("/:id/tags", coursesAPI.createTag)
	courseGroup.Delete("/:id/tags/:tagId", coursesAPI.deleteTag)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const bufferSize = 1024 * 8                 // 8KB per chunk, adjust as needed
const maxInitialChunkSize = 1024 * 1024 * 5 // 5MB, adjust as needed

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getCourses(c *fiber.Ctx) error {
	orderBy := c.Query("orderBy", models.COURSE_TABLE+".created_at desc")
	titles := c.Query("titles", "")
	progress := c.Query("progress", "")
	tags := c.Query("tags", "")

	options := &database.Options{
		OrderBy:    strings.Split(orderBy, ","),
		Pagination: pagination.NewFromApi(c),
	}

	whereClause := squirrel.And{}
	courseIDs := []string{}

	// Filter based on titles
	if titles != "" {
		filtered, err := filter(titles)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid titles parameter", err)
		}

		orClause := squirrel.Or{}
		for _, title := range filtered {
			orClause = append(orClause, squirrel.Like{models.COURSE_TABLE + ".title": "%" + title + "%"})
		}

		whereClause = append(whereClause, orClause)
	}

	// Filter on progress ("not started", "started", "completed") by identifying the course IDs that
	// match the progress filter
	if progress != "" {
		unescapedProgress, err := url.QueryUnescape(progress)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid progress parameter")
		}

		unescapedProgress = strings.ToLower(unescapedProgress)

		switch unescapedProgress {
		case "not started":
			courseIDs, err = api.dao.PluckIDsForNotStartedCourses(c.UserContext(), nil)
		case "started":
			courseIDs, err = api.dao.PluckIDsForStartedCourses(c.UserContext(), nil)
		case "completed":
			courseIDs, err = api.dao.PluckIDsForCompletedCourses(c.UserContext(), nil)
		}

		if err != nil {
			return errorResponse(c, fiber.StatusInternalServerError, "Error looking up courses by progress", err)
		}

		if len(courseIDs) == 0 {
			pResult, err := options.Pagination.BuildResult(courseResponseHelper(nil))
			if err != nil {
				return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
			}

			return c.Status(fiber.StatusOK).JSON(pResult)
		}
	}

	// Filter based on tags by identifying the course IDs that match the tags filter
	if tags != "" {
		filtered, err := filter(tags)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "Invalid tags parameter", err)
		}

		tmpCourseIDs, err := api.dao.PluckCourseIDsWithTags(c.UserContext(), filtered, nil)
		if err != nil {
			return errorResponse(c, fiber.StatusInternalServerError, "Error looking up courses by tags", err)
		}

		if len(tmpCourseIDs) == 0 {
			pResult, err := options.Pagination.BuildResult(courseResponseHelper(nil))
			if err != nil {
				return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
			}

			return c.Status(fiber.StatusOK).JSON(pResult)
		}

		if len(courseIDs) != 0 {
			courseIDs = utils.SliceIntersection(courseIDs, tmpCourseIDs)
		} else {
			courseIDs = tmpCourseIDs
		}
	}

	if len(courseIDs) > 0 {
		whereClause = append(whereClause, squirrel.Eq{models.COURSE_TABLE + ".id": courseIDs})
	}

	options.Where = whereClause

	courses := []*models.Course{}
	err := api.dao.List(c.UserContext(), &courses, options)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up courses", err)
	}

	pResult, err := options.Pagination.BuildResult(courseResponseHelper(courses))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	course := &models.Course{Base: models.Base{ID: id}}
	err := api.dao.GetById(c.UserContext(), course)

	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Course not found", fmt.Errorf("course not found"))
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course", err)
	}

	return c.Status(fiber.StatusOK).JSON(courseResponseHelper([]*models.Course{course})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) createCourse(c *fiber.Ctx) error {
	course := new(models.Course)

	if err := c.BodyParser(course); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	// Ensure there is a title and path
	if course.Title == "" || course.Path == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A title and path are required", nil)
	}

	// Empty stuff that should not be set
	course.ID = ""

	course.Path = utils.NormalizeWindowsDrive(course.Path)

	// Validate the path
	if exists, err := afero.DirExists(api.appFs.Fs, course.Path); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid course path", err)
	}

	// Set the course to available
	course.Available = true

	if err := api.dao.CreateCourse(c.UserContext(), course); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "A course with this path already exists", err)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating course", err)
	}

	// Start a scan job
	if scan, err := api.courseScan.Add(c.UserContext(), course.ID); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error creating scan job", err)
	} else {
		course.ScanStatus = scan.Status
	}

	return c.Status(fiber.StatusCreated).JSON(courseResponseHelper([]*models.Course{course})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) deleteCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	course := &models.Course{Base: models.Base{ID: id}}
	err := api.dao.Delete(c.UserContext(), course, nil)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting course", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getCard(c *fiber.Ctx) error {
	id := c.Params("id")

	course := &models.Course{Base: models.Base{ID: id}}
	err := api.dao.GetById(c.UserContext(), course)

	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Course not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course", err)
	}

	if course.CardPath == "" {
		return errorResponse(c, fiber.StatusNotFound, "Course has no card", nil)
	}

	_, err = api.appFs.Fs.Stat(course.CardPath)
	if os.IsNotExist(err) {
		return errorResponse(c, fiber.StatusNotFound, "Course card not found", nil)
	}

	// The fiber function sendFile(...) does not support using a custom FS. Therefore, use
	// SendFile() from the filesystem middleware.
	return filesystem.SendFile(c, afero.NewHttpFs(api.appFs.Fs), course.CardPath)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getAssets(c *fiber.Ctx) error {
	id := c.Params("id")
	orderBy := c.Query("orderBy", "chapter asc,prefix asc")

	options := &database.Options{
		OrderBy:    strings.Split(orderBy, ","),
		Where:      squirrel.Eq{models.ASSET_TABLE + ".course_id": id},
		Pagination: pagination.NewFromApi(c),
	}

	assets := []*models.Asset{}
	err := api.dao.List(c.UserContext(), &assets, options)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up assets", err)
	}

	pResult, err := options.Pagination.BuildResult(assetResponseHelper(assets))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getAsset(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	return c.Status(fiber.StatusOK).JSON(assetResponseHelper([]*models.Asset{asset})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) serveAsset(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	// Check for invalid path
	if exists, err := afero.Exists(api.appFs.Fs, asset.Path); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not exist", nil)
	}

	if asset.Type.IsVideo() {
		return handleVideo(c, api.appFs, asset)
	} else if asset.Type.IsHTML() {
		return handleHtml(c, api.appFs, asset)
	}

	// TODO: Handle PDF
	return c.Status(fiber.StatusOK).SendString("done")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) updateAssetProgress(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")

	req := &assetProgressRequest{}
	if err := c.BodyParser(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	assetProgress := &models.AssetProgress{
		AssetID:   assetId,
		VideoPos:  req.VideoPos,
		Completed: req.Completed,
	}

	err = api.dao.CreateOrUpdateAssetProgress(c.UserContext(), assetProgress)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error updating asset", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getAttachments(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	orderBy := c.Query("orderBy", "title asc")

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	options := &database.Options{
		OrderBy:    strings.Split(orderBy, ","),
		Where:      squirrel.Eq{models.ATTACHMENT_TABLE + ".asset_id": assetId},
		Pagination: pagination.NewFromApi(c),
	}

	attachments := []*models.Attachment{}
	err = api.dao.List(c.UserContext(), &attachments, options)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachments", err)
	}

	pResult, err := options.Pagination.BuildResult(attachmentResponseHelper(attachments))
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error building pagination result", err)
	}

	return c.Status(fiber.StatusOK).JSON(pResult)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	attachmentId := c.Params("attachment")

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	attachment := &models.Attachment{Base: models.Base{ID: attachmentId}}
	err = api.dao.GetById(c.UserContext(), attachment)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Attachment not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachment", err)
	}

	if attachment.AssetID != assetId {
		return errorResponse(c, fiber.StatusBadRequest, "Attachment does not belong to asset", nil)
	}

	return c.Status(fiber.StatusOK).JSON(attachmentResponseHelper([]*models.Attachment{attachment})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) serveAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	assetId := c.Params("asset")
	attachmentID := c.Params("attachment")

	asset := &models.Asset{Base: models.Base{ID: assetId}}
	err := api.dao.GetById(c.UserContext(), asset)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Asset not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up asset", err)
	}

	if asset.CourseID != id {
		return errorResponse(c, fiber.StatusBadRequest, "Asset does not belong to course", nil)
	}

	attachment := &models.Attachment{Base: models.Base{ID: attachmentID}}
	err = api.dao.GetById(c.UserContext(), attachment)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Attachment not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up attachment", err)
	}

	if attachment.AssetID != assetId {
		return errorResponse(c, fiber.StatusBadRequest, "Attachment does not belong to asset", nil)
	}

	if exists, err := afero.Exists(api.appFs.Fs, attachment.Path); err != nil || !exists {
		return errorResponse(c, fiber.StatusBadRequest, "Attachment does not exist", err)
	}

	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+attachment.Title+`"`)
	return filesystem.SendFile(c, afero.NewHttpFs(api.appFs.Fs), attachment.Path)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) getTags(c *fiber.Ctx) error {
	id := c.Params("id")

	course := &models.Course{Base: models.Base{ID: id}}
	err := api.dao.GetById(c.UserContext(), course)
	if err != nil {
		if err == sql.ErrNoRows {
			return errorResponse(c, fiber.StatusNotFound, "Course not found", nil)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course", err)
	}

	options := &database.Options{
		OrderBy: []string{models.TAG_TABLE + ".tag asc"},
		Where:   squirrel.Eq{models.COURSE_TAG_TABLE + ".course_id": id},
	}

	tags := []*models.CourseTag{}
	err = api.dao.List(c.UserContext(), &tags, options)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error looking up course tags", err)
	}

	return c.Status(fiber.StatusOK).JSON(courseTagResponseHelper(tags))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) createTag(c *fiber.Ctx) error {
	courseId := c.Params("id")
	courseTag := new(models.CourseTag)

	if err := c.BodyParser(courseTag); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Error parsing data", err)
	}

	courseTag.ID = ""
	courseTag.TagID = ""

	if courseTag.Tag == "" {
		return errorResponse(c, fiber.StatusBadRequest, "A tag is required", nil)
	}

	courseTag.CourseID = courseId

	err := api.dao.CreateCourseTag(c.UserContext(), courseTag)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errorResponse(c, fiber.StatusBadRequest, "A tag for this course already exists", err)
		}

		return errorResponse(c, fiber.StatusInternalServerError, "Error creating course tag", err)
	}

	return c.Status(fiber.StatusCreated).JSON(courseTagResponseHelper([]*models.CourseTag{courseTag})[0])
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (api coursesAPI) deleteTag(c *fiber.Ctx) error {
	courseId := c.Params("id")
	tagId := c.Params("tagId")

	err := api.dao.Delete(
		c.UserContext(),
		&models.CourseTag{},
		&database.Options{Where: squirrel.And{squirrel.Eq{"course_id": courseId}, squirrel.Eq{"id": tagId}}},
	)

	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error deleting course tag", err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
