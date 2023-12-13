package app

import (
	"fmt"
	"io"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logFile is the path and filename to store the application's log
const logFileDirName = "log"
const dataStorageLogPath = dataStorageRootPath + "/" + logFileDirName

const logFileBaseName = "lego-certhub"
const logFileSuffix = ".log"
const logFileName = logFileBaseName + logFileSuffix

// defaultLogLevel is the default logging level when not in devMode
// and the configured level isn't valid or specified
const defaultLogLevel = zapcore.InfoLevel

// initZapLogger starts the app's logger. if app has not yet read config, it
// uses some default settings to log the initial app start before a second call
// which then recreates the logger with the configured settings
func (app *Application) initZapLogger() {
	// log level based on devMode and config
	var logLevelParseErr error

	// if config isn't loaded, use debug
	logLevel := zapcore.DebugLevel
	// try to load level from config
	if app.config != nil && app.config.LogLevel != nil {
		logLevel, logLevelParseErr = zapcore.ParseLevel(*app.config.LogLevel)
		// if parse error, set default log level
		if logLevelParseErr != nil {
			logLevel = defaultLogLevel
			// do not log error here, will be logged later after logger is configured
		}
	}

	// file writer
	// make zap config for file
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	// to make the whole file Unmarshal-able
	config.LineEnding = ",\n"

	// log file encoder
	fileEncoder := zapcore.NewJSONEncoder(config)

	// if running for initial app start, do not use lumberjack
	// this is because of lumberjack bug that causes a go routine leak
	// manually log to file for initial start up to avoid this leak
	// this may lead to log not rotating at the exact correct time, but that's
	// fine, as it is a relatively minimal number of log lines for initial start
	var writer interface{ io.Writer }
	useFile := true
	// for closing the underlying log file later
	closeFileFunc := func() error { return nil }

	// if no config (init run) manually open file
	if app.config == nil {
		f, err := openLogFile()

		// confirm file opened, else don't use file
		if f != nil && err == nil {
			writer = f
			closeFileFunc = f.Close
		} else {
			useFile = false
		}

	} else {
		// not init run, make lumber jack
		lumberjackLogger := &lumberjack.Logger{
			Filename: dataStorageLogPath + "/" + logFileName,
			MaxSize:  1,   // megabytes
			MaxAge:   364, // days
			// MaxBackups: 10,
			LocalTime: true,
			Compress:  false,
		}

		writer = lumberjackLogger
		closeFileFunc = lumberjackLogger.Close
	}
	// end file writer

	// console
	config.LineEnding = "\n" // regular line break for console
	// no stack trace in console
	config.StacktraceKey = ""
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	// create logger core for console
	core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel)

	// if file writing, Tee it on
	if useFile {
		// add sync
		writeSyncer := zapcore.AddSync(writer)

		core = zapcore.NewTee(
			zapcore.NewCore(fileEncoder, writeSyncer, logLevel),
			core,
		)
	}

	// flush and close any previous logger before overwriting
	if app.logger != nil {
		app.logger.syncAndClose()
	}

	// make logger - don't change address during re-init
	if app.logger == nil {
		app.logger = &appLogger{}
		app.logger.SugaredLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()
	} else {
		*app.logger.SugaredLogger = *zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()
	}

	app.logger.syncAndClose = func() {
		_ = app.logger.Sync()
		_ = closeFileFunc()
	}

	// log if parsing log level failed earlier
	if logLevelParseErr != nil {
		app.logger.Warnf("failed to parse config log level ('%s' not valid), using default (%s)", *app.config.LogLevel, defaultLogLevel)
	}

	// log which init this is
	if app.config == nil {
		app.logger.Infof("init logging started (log level: %s)", logLevel)
	} else {
		app.logger.Infof("main logging started (log level: %s)", logLevel)
	}
}

// openLogFile opens the app's current log file or creates one if it does not exist
func openLogFile() (*os.File, error) {
	// make log path if it does not exist
	err := os.MkdirAll(dataStorageLogPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("can't make directories for new logfile: %s", err)
	}

	// open log file (create if does not exist)
	f, err := os.OpenFile(dataStorageLogPath+"/"+logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.FileMode(0600))
	if err != nil {
		return nil, fmt.Errorf("failed to open or create logfile: %s", err)
	}

	return f, nil
}
