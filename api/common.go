package api

import (
	"fmt"
	"mime"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/spf13/afero"
)

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
			ScanStatus: course.ScanStatus.String(),

			// Progress
			Progress: courseProgressResponse{
				Started:           course.Progress.Started,
				StartedAt:         course.Progress.StartedAt,
				Percent:           course.Progress.Percent,
				CompletedAt:       course.Progress.CompletedAt,
				ProgressUpdatedAt: course.Progress.UpdatedAt,
			},
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func assetResponseHelper(assets []*models.Asset) []*assetResponse {
	responses := []*assetResponse{}
	for _, asset := range assets {

		progress := &assetProgressResponse{}
		if asset.Progress != nil {
			progress.VideoPos = asset.Progress.VideoPos
			progress.Completed = asset.Progress.Completed
			progress.CompletedAt = asset.Progress.CompletedAt
		}

		responses = append(responses, &assetResponse{
			ID:        asset.ID,
			CourseID:  asset.CourseID,
			Title:     asset.Title,
			Prefix:    int(asset.Prefix.Int16),
			Chapter:   asset.Chapter,
			Path:      asset.Path,
			Type:      asset.Type,
			CreatedAt: asset.CreatedAt,
			UpdatedAt: asset.UpdatedAt,

			Progress:    progress,
			Attachments: attachmentResponseHelper(asset.Attachments),
		})

	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func attachmentResponseHelper(attachments []*models.Attachment) []*attachmentResponse {
	responses := []*attachmentResponse{}
	for _, attachment := range attachments {
		responses = append(responses, &attachmentResponse{
			ID:        attachment.ID,
			AssetId:   attachment.AssetID,
			Title:     attachment.Title,
			Path:      attachment.Path,
			CreatedAt: attachment.CreatedAt,
			UpdatedAt: attachment.UpdatedAt,
		})
	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func scanResponseHelper(scans []*models.Scan) []*scanResponse {
	responses := []*scanResponse{}
	for _, scan := range scans {
		responses = append(responses, &scanResponse{
			ID:        scan.ID,
			CourseID:  scan.CourseID,
			Status:    scan.Status,
			CreatedAt: scan.CreatedAt,
			UpdatedAt: scan.UpdatedAt,
		})
	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func tagResponseHelper(tags []*models.Tag) []*tagResponse {
	responses := []*tagResponse{}

	for _, tag := range tags {
		t := &tagResponse{
			ID:          tag.ID,
			Tag:         tag.Tag,
			CreatedAt:   tag.CreatedAt,
			UpdatedAt:   tag.UpdatedAt,
			CourseCount: len(tag.CourseTags),
		}

		// Add the course tags
		if len(tag.CourseTags) > 0 {
			courses := []*courseTagResponse{}

			for _, ct := range tag.CourseTags {
				courses = append(courses, &courseTagResponse{
					ID:       ct.ID,
					CourseID: ct.CourseID,
					Title:    ct.Course,
				})
			}

			t.Courses = courses
		}

		responses = append(responses, t)
	}

	return responses
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// handleVideo handles the video streaming logic
func handleVideo(c *fiber.Ctx, appFs *appFs.AppFs, asset *models.Asset) error {
	// Open the video
	file, err := appFs.Fs.Open(asset.Path)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error opening file", err)
	}
	defer file.Close()

	// Get the file info
	fileInfo, err := file.Stat()
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error getting file info", err)
	}

	// Get the range header and return the entire video if there is no range header
	rangeHeader := c.Get("Range", "")
	if rangeHeader == "" {
		return filesystem.SendFile(c, afero.NewHttpFs(appFs.Fs), asset.Path)
	}

	// Parse the "bytes=START-END" format
	bytesPos := strings.Split(rangeHeader, "=")[1]
	rangeStartEnd := strings.Split(bytesPos, "-")
	start, _ := strconv.Atoi(rangeStartEnd[0])
	var end int

	if len(rangeStartEnd) == 2 && rangeStartEnd[0] == "0" && rangeStartEnd[1] == "1" {
		start = 0
		end = 1
	} else {
		end = start + maxInitialChunkSize - 1
		if end >= int(fileInfo.Size()) {
			end = int(fileInfo.Size()) - 1
		}
	}

	if start > end {
		return errorResponse(c, fiber.StatusBadRequest, "Range start cannot be greater than end", fmt.Errorf("range start is greater than end"))
	}

	// Setting required response headers
	c.Set(fiber.HeaderContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
	c.Set(fiber.HeaderContentLength, strconv.Itoa(end-start+1))
	c.Set(fiber.HeaderContentType, mime.TypeByExtension(filepath.Ext(asset.Path)))
	c.Set(fiber.HeaderAcceptRanges, "bytes")

	// Set the status code to 206 Partial Content
	c.Status(fiber.StatusPartialContent)

	file.Seek(int64(start), 0)
	buffer := make([]byte, bufferSize)
	bytesToSend := end - start + 1

	// Respond in chunks
	for bytesToSend > 0 {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			break
		}

		if bytesRead > bytesToSend {
			bytesRead = bytesToSend
		}

		c.Write(buffer[:bytesRead])
		bytesToSend -= bytesRead
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// handleHtml handles serving HTML files
func handleHtml(c *fiber.Ctx, appFs *appFs.AppFs, asset *models.Asset) error {
	// Open the HTML file
	file, err := appFs.Fs.Open(asset.Path)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error opening file", err)
	}
	defer file.Close()

	// Read the content of the HTML file
	content, err := afero.ReadAll(file)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Error reading file", err)
	}

	c.Set(fiber.HeaderContentType, "text/html")
	return c.Status(fiber.StatusOK).Send(content)
}
