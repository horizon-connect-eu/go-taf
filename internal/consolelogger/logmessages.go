package consolelogger

import "github.com/pterm/pterm"

type logOutput struct {
	message string
	level   pterm.LogLevel
	//	args    map[string]any
}

type logTable struct {
	table [][]string
}
