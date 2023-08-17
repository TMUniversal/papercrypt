package internal

import (
	"log"
	"os"
)

// L is the log writer for regular output
var L = log.New(os.Stderr, "", 0)
