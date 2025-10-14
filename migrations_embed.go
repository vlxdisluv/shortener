package shortener

import (
	"embed"
)

//go:embed db/migrations/*.sql
var FS embed.FS
