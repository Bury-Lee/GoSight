package global

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger
var LogFile *os.File
