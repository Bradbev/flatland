package asset

import (
	systemLog "log"
	"os"
)

var log = systemLog.New(os.Stderr, "Asset", systemLog.Ltime)
