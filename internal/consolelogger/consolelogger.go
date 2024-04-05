package consolelogger

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"
	"time"
)

func NewLogger() Logger {
	return Logger{
		logChan:   make(chan any, 100),
		startTime: time.Now(),
	}
}

type Logger struct {
	logChan   chan any
	tableData pterm.TableData
	startTime time.Time
}

func (l *Logger) Run(ctx context.Context) {
	clearTerminal()
	//logger := pterm.DefaultLogger.WithLevel(pterm.LogLevelTrace).WithTime(true).WithTimeFormat("04:05.000")
	logger := pterm.DefaultLogger.WithLevel(pterm.LogLevelTrace).WithTime(false)

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(ctx); err != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case msg := <-l.logChan:

			deltaTime := time.Now().Sub(l.startTime)
			timestamp := fmt.Sprintf("[%s]", deltaTime.Round(time.Millisecond))

			switch logMsg := msg.(type) {
			case logOutput:
				switch logMsg.level {
				case pterm.LogLevelInfo:
					logger.Info(timestamp+" "+logMsg.message, logMsg.args)
				case pterm.LogLevelWarn:
					logger.Warn(timestamp+" "+logMsg.message, logMsg.args)
				}
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

func (l *Logger) InfoWithArgs(msg string, args ...pterm.LoggerArgument) {
	l.logChan <- logOutput{
		message: msg,
		level:   pterm.LogLevelInfo,
		args:    args,
	}
}

func (l *Logger) Warn(msg string) {
	l.logChan <- logOutput{
		message: msg,
		level:   pterm.LogLevelWarn,
	}
}
func (l *Logger) WarnWithArgs(msg string, args ...pterm.LoggerArgument) {
	l.logChan <- logOutput{
		message: msg,
		level:   pterm.LogLevelWarn,
		args:    args,
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
