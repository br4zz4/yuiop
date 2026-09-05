// Package yuiop exposes build-time embedded assets.
package yuiop

import _ "embed"

//go:embed data/packages.yml
var PackagesYAML []byte
