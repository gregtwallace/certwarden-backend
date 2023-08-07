package dns01acmesh

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"runtime"

	"go.uber.org/zap"
)

const (
	acmeShFileName = "acme.sh"
	dnsApiPath     = "/dnsapi"
	tempScriptPath = "/temp"
)

var (
	errServiceComponent = errors.New("necessary dns-01 acme.sh component is missing")
	errNoAcmeShPath     = errors.New("acme.sh path not specified in config")
	errBashMissing      = errors.New("unable to find bash")
	errWindows          = errors.New("acme.sh is not supported in windows, disable it")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
}

// Accounts service struct
type Service struct {
	logger          *zap.SugaredLogger
	shellPath       string
	shellScriptPath string
	dnsHook         string
	environmentVars []string
}

// Configuration options
type Config struct {
	Enable      *bool    `yaml:"enable"`
	AcmeShPath  *string  `yaml:"acme_sh_path"`
	Environment []string `yaml:"environment"`
	DnsHook     string   `yaml:"dns_hook"`
}

// NewService creates a new service
func NewService(app App, config *Config) (*Service, error) {
	var err error

	// if disabled, return nil and no error
	if !*config.Enable {
		return nil, nil
	}

	// add error and fail for trying to run on windows
	if runtime.GOOS == "windows" {
		return nil, errWindows
	}

	// error if no path
	if config.AcmeShPath == nil {
		return nil, errNoAcmeShPath
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// bash is required
	service.shellPath, err = exec.LookPath("bash")
	if err != nil {
		return nil, errBashMissing
	}

	// read in base script
	acmeSh, err := os.ReadFile(*config.AcmeShPath + "/" + acmeShFileName)
	if err != nil {
		return nil, err
	}
	// remove execution of main func (`main "$@"`)
	acmeSh, _, _ = bytes.Cut(acmeSh, []byte{109, 97, 105, 110, 32, 34, 36, 64, 34})

	// read in dns_hook script
	acmeShDnsHook, err := os.ReadFile(*config.AcmeShPath + dnsApiPath + "/" + config.DnsHook + ".sh")
	if err != nil {
		return nil, err
	}

	// combine scripts
	shellScript := append(acmeSh, acmeShDnsHook...)

	// store in file to use as source
	path := *config.AcmeShPath + tempScriptPath
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	service.shellScriptPath = path + "/" + acmeShFileName + "_" + config.DnsHook + ".sh"

	shellFile, err := os.Create(service.shellScriptPath)
	if err != nil {
		return nil, err
	}
	defer shellFile.Close()

	_, err = shellFile.Write(shellScript)
	if err != nil {
		return nil, err
	}

	// hook name (needed for funcs)
	service.dnsHook = config.DnsHook

	// environment vars
	service.environmentVars = config.Environment

	return service, nil
}
