package consolelogger

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"
)

func NewLogger() Logger {
	return Logger{
		logChan: make(chan any, 100),
	}
}

type Logger struct {
	logChan   chan any
	tableData pterm.TableData
}

func (l *Logger) Run(ctx context.Context) {
	clearTerminal()
	logger := pterm.DefaultLogger.WithLevel(pterm.LogLevelTrace)

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(ctx); err != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case msg := <-l.logChan:

			switch logMsg := msg.(type) {
			case logOutput:
				logger.WithLevel(logMsg.level).Print(logMsg.message)
			case logTable:
				l.tableData = pterm.TableData(logMsg.table)
				pterm.DefaultTable.WithHasHeader().WithBoxed().WithRightAlignment().WithData(l.tableData).Render()
			}
		}
	}
}

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
}

func (l *Logger) Info(msg string) {
	l.logChan <- logOutput{
		message: msg,
		level:   pterm.LogLevelInfo,
	}
}

func (l *Logger) Warn(msg string) {
	l.logChan <- logOutput{
		message: msg,
		level:   pterm.LogLevelWarn,
	}
}
func (l *Logger) Error(msg string) {
	l.logChan <- logOutput{
		message: msg,
		level:   pterm.LogLevelError,
	}
}

func (l *Logger) Table(table [][]string) {
	l.logChan <- logTable{
		table: table,
	}
}
