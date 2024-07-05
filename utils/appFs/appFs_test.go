package appFs

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/geerew/off-course/utils/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T) (*AppFs, *[]*logger.Log) {
	t.Helper()

	// Logger
	var logs []*logger.Log
	var logsMux sync.Mutex
	logger, _, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize: 1,
		WriteFn:   logger.TestWriteFn(&logs, &logsMux),
	})
	require.NoError(t, err, "Failed to initialize logger")

	// Filesystem
	appFs := NewAppFs(afero.NewMemMapFs(), logger)

	return appFs, &logs
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Open(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		appFs, _ := setup(t)

		res, err := appFs.Open("'")

		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
		require.Nil(t, res)
	})

	t.Run("file exists", func(t *testing.T) {
		appFs, _ := setup(t)

		appFs.Fs.Create("/a")

		res, err := appFs.Open("/a")
		require.Nil(t, err)
		require.NotNil(t, res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ReadDir(t *testing.T) {
	t.Run("open error", func(t *testing.T) {
		appFs, logs := setup(t)

		res, err := appFs.ReadDir("'", false)

		require.Nil(t, res)
		require.EqualError(t, err, "unable to open path")
		require.Len(t, *logs, 1)
		require.Equal(t, "Unable to open path", (*logs)[0].Message)
		require.Equal(t, slog.LevelError, (*logs)[0].Level)
	})

	t.Run("read path error", func(t *testing.T) {
		appFs, logs := setup(t)

		appFs.Fs.Create("/test")
		res, err := appFs.ReadDir("/test", false)

		require.Nil(t, res)
		require.EqualError(t, err, "unable to read path")
		require.Len(t, *logs, 1)
		require.Equal(t, "Unable to read path", (*logs)[0].Message)
		require.Equal(t, slog.LevelError, (*logs)[0].Level)
	})

	t.Run("success", func(t *testing.T) {
		appFs, _ := setup(t)

		appFs.Fs.Create("/a")
		appFs.Fs.Create("/b")
		appFs.Fs.Mkdir("/c", 0755)

		res, err := appFs.ReadDir("/", true)
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, 2, len(res.Files))
		require.Equal(t, 1, len(res.Directories))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ReadDirFlat(t *testing.T) {
	t.Run("open error", func(t *testing.T) {
		appFs, logs := setup(t)

		res, err := appFs.ReadDirFlat("'", 1)

		require.Nil(t, res)
		require.EqualError(t, err, "unable to open path")
		require.Len(t, *logs, 1)
		require.Equal(t, "Unable to open path", (*logs)[0].Message)
		require.Equal(t, slog.LevelError, (*logs)[0].Level)
	})

	t.Run("read path error", func(t *testing.T) {
		appFs, logs := setup(t)

		appFs.Fs.Create("/test")
		res, err := appFs.ReadDirFlat("/test", 1)

		require.Nil(t, res)
		require.EqualError(t, err, "unable to read path")
		require.Len(t, *logs, 1)
		require.Equal(t, "Unable to read path", (*logs)[0].Message)
		require.Equal(t, slog.LevelError, (*logs)[0].Level)
	})

	t.Run("success", func(t *testing.T) {
		appFs, _ := setup(t)

		// Top files/dirs
		appFs.Fs.Mkdir("/1", 0755)
		appFs.Fs.Mkdir("/2/2", 0755)
		appFs.Fs.Mkdir("/3/3/3", 0755)
		appFs.Fs.Mkdir("/4/4/4/4", 0755)
		appFs.Fs.Create("/f1")
		appFs.Fs.Create("/1/f1")
		appFs.Fs.Create("/2/f1")
		appFs.Fs.Create("/2/2/f1")
		appFs.Fs.Create("/3/f1")
		appFs.Fs.Create("/3/3/f1")
		appFs.Fs.Create("/3/3/3/f1")
		appFs.Fs.Create("/4/f1")
		appFs.Fs.Create("/4/4/f1")
		appFs.Fs.Create("/4/4/4/f1")
		appFs.Fs.Create("/4/4/4/4/f1")

		// Depth 0 (same as 1)
		res, err := appFs.ReadDirFlat("/", 0)
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, 1, len(res))

		// Depth 1
		res, err = appFs.ReadDirFlat("/", 1)
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, 1, len(res))

		// Depth 10
		res, err = appFs.ReadDirFlat("/", 2)
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, 5, len(res))

		// Depth 10
		res, err = appFs.ReadDirFlat("/", 10)
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, 11, len(res))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NonWslDrives(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		appFs, _ := setup(t)

		// Create WSL directory structure
		paths := []string{}
		for _, p := range paths {
			err := appFs.Fs.MkdirAll(p, os.ModePerm)
			require.Nil(t, err)
		}

		drives, err := appFs.nonWslDrives()

		if errors.Is(err, fmt.Errorf("not implemented yet")) {
			t.Skip("not implemented")
		}

		require.Nil(t, err)
		require.NotEmpty(t, drives)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_WslDrives(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		appFs, logs := setup(t)

		paths := []string{}
		for _, p := range paths {
			err := appFs.Fs.MkdirAll(p, os.ModePerm)
			require.Nil(t, err)
		}

		drives, err := appFs.wslDrives()

		require.Nil(t, drives)
		require.EqualError(t, err, "unable to open path")
		require.Len(t, *logs, 1)
		require.Equal(t, "Unable to open path", (*logs)[0].Message)
		require.Equal(t, slog.LevelError, (*logs)[0].Level)
	})

	t.Run("success", func(t *testing.T) {
		appFs, _ := setup(t)

		// Create WSL directory structure
		paths := []string{"/mnt/c", "/mnt/d", "/mnt/wsl", "/mnt/wslg"}
		for _, p := range paths {
			err := appFs.Fs.MkdirAll(p, os.ModePerm)
			require.Nil(t, err)
		}

		drives, err := appFs.wslDrives()
		require.Nil(t, err)
		require.Len(t, drives, 3)
		require.ElementsMatch(t, []string{"/", filepath.Join("/mnt", "c"), filepath.Join("/mnt", "d")}, drives)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_PartialHash(t *testing.T) {
	t.Run("random sizes", func(t *testing.T) {
		appFs, _ := setup(t)

		bigData := make([]byte, 1024*1024*10)
		for i := 0; i < len(bigData); i++ {
			bigData[i] = byte(i % 256)
		}

		tests := []struct {
			name     string
			content  []byte
			expected string // You can generate these expected hashes beforehand
		}{
			{"empty", []byte(""), "af5570f5a1810b7af78caf4bc70a660f0df51e42baf91d4de5b2328de0e83dfc"},
			{"small", []byte("small file contents"), "db664be16228614363fb0506a9f828fdb0dbb5ceef6465ac344647ed6feae240"},
			{"big", bigData, "c37339299959bdb2885f2f30f77247b0fa3760ce1181d418b4f17fa652ff1386"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				require.Nil(t, afero.WriteFile(appFs.Fs, "/test/"+tt.name, tt.content, 0644))

				hash, err := appFs.PartialHash("/test/"+tt.name, 1024*1024)
				require.Nil(t, err)
				require.Equal(t, tt.expected, hash)
			})
		}
	})

	t.Run("rename file", func(t *testing.T) {
		appFs, _ := setup(t)

		require.Nil(t, afero.WriteFile(appFs.Fs, "/test/data", []byte("Some test data"), 0644))

		hash, err := appFs.PartialHash("/test/data", 1024*1024)
		require.Nil(t, err)
		require.Equal(t, "0843f7816915fae7fc9c31dbbb3e8745015b53a297930e522d544c13287cb062", hash)

		// Rename the file
		require.Nil(t, appFs.Fs.Rename("/test/data", "/test/newdata"))

		hash, err = appFs.PartialHash("/test/newdata", 1024*1024)
		require.Nil(t, err)
		require.Equal(t, "0843f7816915fae7fc9c31dbbb3e8745015b53a297930e522d544c13287cb062", hash)
	})
}
