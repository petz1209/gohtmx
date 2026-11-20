package home

import "embed"

//go:embed templates static
var files embed.FS

// assets.go macht nur die referenzen auf die emedded Filesystems.
// in diesem fall wird sowohl templates als auch static embedded.
// das is notwendig um das feature sauber einzubinden.
