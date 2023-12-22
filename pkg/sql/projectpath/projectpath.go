package projectpath

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// Root is the root of the project
	Root = filepath.Join(filepath.Dir(b), "../../..")
)
