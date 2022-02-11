package chart

import _ "embed"

//go:embed bundle.js
var bundleJs []byte

//go:embed bundle.js.LICENSE.txt
var licenseTxt []byte

//go:embed index.html
var indexHtml []byte
