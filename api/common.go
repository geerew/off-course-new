package api

import (
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// errorResponse is a helper method to return an error response
func errorResponse(c *fiber.Ctx, status int, message string, err error) error {
	resp := fiber.Map{
		"message": message,
	}

	if err != nil {
		resp["error"] = err.Error()
	}

	return c.Status(status).JSON(resp)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func filter(s string) ([]string, error) {
	unescaped, err := url.QueryUnescape(s)
	if err != nil {
		return nil, err
	}

	return utils.Map(strings.Split(unescaped, ","), strings.TrimSpace), nil
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
