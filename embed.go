package embedded

import "embed"

//go:embed  res/schemas/*
var Schemas embed.FS
