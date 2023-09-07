package dns01acmesh

import (
	"bytes"
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
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

// provider Service struct
type Service struct {
	logger          *zap.SugaredLogger
	domains         []string
	shellPath       string
	shellScriptPath string
	dnsHook         string
	environmentVars []string
}

// Stop/Start is not needed for this provider. Nothing needs to be stopped or started.
func (service *Service) Stop() error  { return nil }
func (service *Service) Start() error { return nil }

// Configuration options
type Config struct {
	Doms        []string                         `yaml:"domains" json:"domains"`
	AcmeShPath  *string                          `yaml:"acme_sh_path" json:"acme_sh_path"`
	Environment output.RedactedEnvironmentParams `yaml:"environment" json:"environment"`
	DnsHook     string                           `yaml:"dns_hook" json:"dns_hook"`
}

// Domains returns all of the domains specified in the Config
func (cfg *Config) Domains() []string {
	return cfg.Doms
}

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
	// error and fail if trying to run on windows
	if runtime.GOOS == "windows" {
		return nil, errWindows
	}

	// if no config or no domains, error
	if cfg == nil || len(cfg.Doms) <= 0 {
		return nil, errServiceComponent
	}

	// error if no path
	if cfg.AcmeShPath == nil {
		return nil, errNoAcmeShPath
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// set supported domains from config
	service.domains = append(service.domains, cfg.Doms...)

	// bash is required
	var err error
	service.shellPath, err = exec.LookPath("bash")
	if err != nil {
		return nil, errBashMissing
	}

	// read in base script
	acmeSh, err := os.ReadFile(*cfg.AcmeShPath + "/" + acmeShFileName)
	if err != nil {
		return nil, err
	}
	// remove execution of main func (`main "$@"`)
	acmeSh, _, _ = bytes.Cut(acmeSh, []byte{109, 97, 105, 110, 32, 34, 36, 64, 34})

	// read in dns_hook script
	acmeShDnsHook, err := os.ReadFile(*cfg.AcmeShPath + dnsApiPath + "/" + cfg.DnsHook + ".sh")
	if err != nil {
		return nil, err
	}

	// combine scripts
	shellScript := append(acmeSh, acmeShDnsHook...)

	// store in file to use as source
	path := *cfg.AcmeShPath + tempScriptPath
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	service.shellScriptPath = path + "/" + acmeShFileName + "_" + cfg.DnsHook + ".sh"

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
	service.dnsHook = cfg.DnsHook

	// environment vars
	service.environmentVars = cfg.Environment.Unredacted()

	return service, nil
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDns01
}
