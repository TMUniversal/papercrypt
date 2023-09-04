package internal

import (
	"github.com/pkg/errors"
	"log"
	"os"
)

var l = log.New(os.Stderr, "", 0)

func GetFileHandleCarefully(path string, override bool) (*os.File, error) {
	if path == "" || path == "-" {
		return os.Stdout, nil
	}

	if _, err := os.Stat(path); err == nil {
		if !override {
			return nil, errors.Errorf("file %s already exists, use --force to override", path)
		} else {
			l.Printf("Overriding existing file \"%s\"!\n", path)
		}
	}

	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, errors.Errorf("error opening file '%s': %s", path, err)
	}

	return out, nil
}
