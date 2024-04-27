package appFs

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rzajac/zltest"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Open(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		// Create fs and read from invalid path
		appFs := NewAppFs(afero.NewMemMapFs())
		res, err := appFs.Open("'")

		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
		require.Nil(t, res)
	})

	t.Run("file exists", func(t *testing.T) {
		appFs := NewAppFs(afero.NewMemMapFs())

		appFs.Fs.Create("/a")

		res, err := appFs.Open("/a")
		require.Nil(t, err)
		require.NotNil(t, res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ReadDir(t *testing.T) {
	// Set logger
	logHook := zltest.New(t)
	log.Logger = zerolog.New(logHook)

	t.Run("open error", func(t *testing.T) {
		// Create fs and read from invalid path
		appFs := NewAppFs(afero.NewMemMapFs())
		res, err := appFs.ReadDir("'", false)

		require.Nil(t, res)
		require.EqualError(t, err, "unable to open path")
		logHook.LastEntry().ExpMsg("unable to open path")
		logHook.LastEntry().ExpLevel(zerolog.ErrorLevel)
	})

	t.Run("read path error", func(t *testing.T) {
		// Create the fs and a file
		appFs := NewAppFs(afero.NewMemMapFs())
		appFs.Fs.Create("/test")

		// Read file as if it's a directory
		res, err := appFs.ReadDir("/test", false)
		require.Nil(t, res)
		require.EqualError(t, err, "unable to read path")

		logHook.LastEntry().ExpMsg("unable to read path")
		logHook.LastEntry().ExpLevel(zerolog.ErrorLevel)
	})

	t.Run("success", func(t *testing.T) {
		// Create fs and add files and directories
		appFs := NewAppFs(afero.NewMemMapFs())

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

func Test_ReadDirFlat(t *testing.T) {
	// Set logger
	logHook := zltest.New(t)
	log.Logger = zerolog.New(logHook)

	t.Run("open error", func(t *testing.T) {
		// Create fs and read from invalid path
		appFs := NewAppFs(afero.NewMemMapFs())
		res, err := appFs.ReadDirFlat("'", 1)

		require.Nil(t, res)
		require.EqualError(t, err, "unable to open path")
		logHook.LastEntry().ExpMsg("unable to open path")
		logHook.LastEntry().ExpLevel(zerolog.ErrorLevel)
	})

	t.Run("read path error", func(t *testing.T) {
		// Create the fs and a file
		appFs := NewAppFs(afero.NewMemMapFs())
		appFs.Fs.Create("/test")

		// Read file as if it's a directory
		res, err := appFs.ReadDirFlat("/test", 1)
		require.Nil(t, res)
		require.EqualError(t, err, "unable to read path")

		logHook.LastEntry().ExpMsg("unable to read path")
		logHook.LastEntry().ExpLevel(zerolog.ErrorLevel)
	})

	t.Run("success", func(t *testing.T) {
		// Create fs and add files and directories
		appFs := NewAppFs(afero.NewMemMapFs())

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
	// Set logger
	logHook := zltest.New(t)
	log.Logger = zerolog.New(logHook)

	t.Run("error", func(t *testing.T) {

		appFs := NewAppFs(afero.NewMemMapFs())

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
		// require.EqualError(t, err, "unable to open path")
	})

	t.Run("success", func(t *testing.T) {
		appFs := NewAppFs(afero.NewMemMapFs())

		// Create WSL directory structure
		paths := []string{"/mnt/c", "/mnt/d", "/mnt/wsl", "/mnt/wslg"}
		for _, p := range paths {
			err := appFs.Fs.MkdirAll(p, os.ModePerm)
			require.Nil(t, err)
		}

		drives, err := appFs.wslDrives()
		require.Nil(t, err)
		require.Len(t, drives, 3)
		require.ElementsMatch(t, []string{"/", "/mnt/c", "/mnt/d"}, drives)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_WslDrives(t *testing.T) {
	// Set logger
	logHook := zltest.New(t)
	log.Logger = zerolog.New(logHook)

	t.Run("error", func(t *testing.T) {
		appFs := NewAppFs(afero.NewMemMapFs())

		// Create WSL directory structure
		paths := []string{}
		for _, p := range paths {
			err := appFs.Fs.MkdirAll(p, os.ModePerm)
			require.Nil(t, err)
		}

		drives, err := appFs.wslDrives()
		require.Nil(t, drives)
		require.EqualError(t, err, "unable to open path")
	})

	t.Run("success", func(t *testing.T) {
		appFs := NewAppFs(afero.NewMemMapFs())

		// Create WSL directory structure
		paths := []string{"/mnt/c", "/mnt/d", "/mnt/wsl", "/mnt/wslg"}
		for _, p := range paths {
			err := appFs.Fs.MkdirAll(p, os.ModePerm)
			require.Nil(t, err)
		}

		drives, err := appFs.wslDrives()
		require.Nil(t, err)
		require.Len(t, drives, 3)
		require.ElementsMatch(t, []string{"/", "/mnt/c", "/mnt/d"}, drives)
	})
}
