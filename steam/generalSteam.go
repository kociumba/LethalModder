package steam

import "errors"

var (
	errNoGame   = errors.New("game not found")
	errVDFParse = errors.New("failed to parse libraryfolders.vdf")

	name = "Lethal Company"
)
