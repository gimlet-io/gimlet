package web

import _ "embed"

//go:embed bundle.js
var BundleJs []byte

//go:embed bundle.js.LICENSE.txt
var LicenseTxt []byte

//go:embed index.html
var IndexHtml []byte
