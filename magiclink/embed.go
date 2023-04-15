package magiclink

import (
	_ "embed"
)

//go:embed frontend/recaptchav3.gohtml
var recaptchav3Template string

//go:embed frontend/default.css
var defaultCSS string
