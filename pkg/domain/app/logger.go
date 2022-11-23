package app

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logFile is the path and filename to store the application's log
const logFile = "./lego-certhub.log.json"

// defaultLogLevel is the default logging level when not in devMode
// and the configured level isn't valid or specified
const defaultLogLevel = zapcore.InfoLevel

func (app *Application) initZapLogger() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	// to make the whole file Unmarshal-able
	config.LineEnding = ",\n"

	// log file
	fileEncoder := zapcore.NewJSONEncoder(config)

	logFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalln(err)
	}

	writer := zapcore.AddSync(logFile)

	// log console
	config.LineEnding = "\n" // regular line break for console
	// no stack trace in console
	config.StacktraceKey = ""
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	// log level based on devMode and config
	logLevel := defaultLogLevel
	var logLevelParseErr error
	// devMode
	if *app.config.DevMode {
		logLevel = zapcore.DebugLevel
	} else {
		// non-dev mode
		logLevel, logLevelParseErr = zapcore.ParseLevel(*app.config.LogLevel)
		// no error check (failed to parse will leave logLevel as default)
	}

	// create logger
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, logLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel),
	)
	app.logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()

	// log if parsing log level failed earlier
	if logLevelParseErr != nil {
		app.logger.Warnf("failed to parse config log level ('%s' not valid)", *app.config.LogLevel)
	}

	app.logger.Infof("logging started (log level: %s)", logLevel)
}
