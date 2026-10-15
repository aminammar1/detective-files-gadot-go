// Package content exposes the embedded case files. The JSON itself is
// authoritative and must not be edited; this shim only makes it importable.
package content

import "embed"

//go:embed cases
var FS embed.FS
