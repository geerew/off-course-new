package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:build
var ui embed.FS

func Assets() fs.FS {
	subFs, err := fs.Sub(ui, "build")

	if err != nil {
		panic(err)
	}

	return subFs
}
