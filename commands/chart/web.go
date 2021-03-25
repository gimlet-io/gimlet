package chart

import _ "embed"

//go:embed bundle.js
var bundlejs []byte

//go:embed bundle.js.LICENSE.txt
var licensetxt []byte

//go:embed index.html
var indexhtml []byte
