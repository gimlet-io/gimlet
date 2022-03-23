package web

import _ "embed"

//go:embed index.html
var IndexHtml []byte

//go:embed step-2.html
var Step2Html []byte

//go:embed step-3.html
var Step3Html []byte

//go:embed server.crt
var ServerCrt []byte

//go:embed server.key
var ServerKey []byte
