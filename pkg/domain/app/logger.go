package app

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (app *Application) initZapLogger() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	// log file
	fileEncoder := zapcore.NewJSONEncoder(config)

	logFile, err := os.OpenFile("lego-certhub.log.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	writer := zapcore.AddSync(logFile)

	// log console
	// no stack trace in console
	config.StacktraceKey = ""
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	// create logger
	// default log level based on dev mode or not
	logLevel := zapcore.InfoLevel
	if *app.config.DevMode {
		logLevel = zapcore.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, logLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel),
	)

	app.logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()
}
