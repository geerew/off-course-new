package api

import (
	"database/sql"
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// Storage interface that is implemented by storage providers
type Storage struct {
	db         *sql.DB
	table      string
	gcInterval time.Duration
	done       chan struct{}
}

// New creates a new storage
func NewSqliteStorage(db *sql.DB, table string) *Storage {
	store := &Storage{
		db:         db,
		table:      table,
		gcInterval: 10 * time.Second,
		done:       make(chan struct{}),
		// sqlSelect:  "SELECT v, e FROM sessions WHERE k=?;",
		// sqlInsert:  "INSERT OR REPLACE INTO sessions (k, v, e) VALUES (?,?,?)",
	}

	// Start garbage collector
	go store.gcTicker()

	return store
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get value by key
func (s *Storage) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, nil
	}

	query := fmt.Sprintf("SELECT data, expires FROM %s WHERE id=?", s.table)
	row := s.db.QueryRow(query, key)

	// Add db response to data
	var (
		data       = []byte{}
		exp  int64 = 0
	)

	if err := row.Scan(&data, &exp); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// If the expiration time has already passed, then return nil
	if exp != 0 && exp <= time.Now().Unix() {
		return nil, nil
	}

	return data, nil
}

// Set key with value
func (s *Storage) Set(key string, data []byte, exp time.Duration) error {
	if key == "" || len(data) <= 0 {
		return nil
	}

	var expSeconds int64
	if exp != 0 {
		expSeconds = time.Now().Add(exp).Unix()
	}

	query := fmt.Sprintf("INSERT OR REPLACE INTO %s (id, data, expires) VALUES (?,?,?)", s.table)
	_, err := s.db.Exec(query, key, data, expSeconds)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes an entry by ID (key)
func (s *Storage) Delete(key string) error {
	if key == "" {
		return nil
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id=?", s.table)
	_, err := s.db.Exec(query, key)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Reset resets all entries, including unexpired
func (s *Storage) Reset() error {
	query := fmt.Sprintf("DELETE FROM %s", s.table)
	_, err := s.db.Exec(query)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Close closes the database
func (s *Storage) Close() error {
	s.done <- struct{}{}
	return s.db.Close()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Conn returns the database client
func (s *Storage) Conn() *sql.DB {
	return s.db
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// gcTicker starts the gc ticker
func (s *Storage) gcTicker() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case t := <-ticker.C:
			query := fmt.Sprintf("DELETE FROM %s WHERE expires <= ? AND expires != 0", s.table)
			_, _ = s.db.Exec(query, t.Unix())
		}
	}
}
