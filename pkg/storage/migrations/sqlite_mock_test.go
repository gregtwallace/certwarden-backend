package migrations_test

import (
	"testing"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const tempFileStorage = "../../../test_data/tmp/"

type gooseLogger struct {
	logger *zap.SugaredLogger
}

// implement goose.Logger
func (gl *gooseLogger) Fatalf(format string, v ...any) {
	gl.logger.Fatalf(format, v)
}

func (gl *gooseLogger) Printf(format string, v ...any) {
	gl.logger.Infof(format, v)
}

// make app
type fakeApp struct {
	gooseLogger *gooseLogger
	dataPath    string
}

func (fa *fakeApp) GetLogger() *zap.SugaredLogger {
	return fa.gooseLogger.logger
}

func (fa *fakeApp) GetDataStorageAppDataPath() string {
	return fa.dataPath
}

func newFakeApp(t *testing.T, dataPath string) *fakeApp {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.FatalLevel)).Sugar() // use fatal to avoid log output

	gl := &gooseLogger{
		logger: logger,
	}

	goose.SetLogger(gl)

	return &fakeApp{
		gooseLogger: gl,
		dataPath:    dataPath,
	}
}
