package web

import _ "embed"

//go:embed main.js
var MainJs []byte

//go:embed main.css
var MainCSS []byte

//go:embed index.html
var IndexHtml []byte

//go:embed 1.chunk.js
var ChunkJs []byte

//go:embed favicon.ico
var Favicon []byte

//go:embed server.crt
var ServerCrt []byte

//go:embed server.key
var ServerKey []byte
