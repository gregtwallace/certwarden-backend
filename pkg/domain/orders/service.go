package orders

import (
	"certwarden-backend/pkg/datatypes/job_manager"
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/domain/authorizations"
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/httpclient"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/pagination_sort"
	"context"
	"errors"
	"os/exec"
	"sync"
	"time"

	"github.com/scaleway/scaleway-sdk-go/logger"
	"go.uber.org/zap"
)

var errServiceComponent = errors.New("orders: necessary service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetShutdownContext() context.Context
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetOrderStorage() Storage
	GetAcmeServerService() *acme_servers.Service
	GetCertificatesService() *certificates.Service

	// for fulfiller
	GetAuthsService() *authorizations.Service
	GetShutdownWaitGroup() *sync.WaitGroup

	IsHttps() bool
	GetHttpClient() *httpclient.Client
	HttpsCertificateName() *string
	LoadHttpsCertificate() error
}

// Storage interface for storage functions
type Storage interface {
	// orders
	GetOneOrder(orderId int) (order Order, err error)
	GetOrders(orderIDs []int) (orders []Order, err error)
	GetOrdersByCert(certId int, q pagination_sort.Query) (orders []Order, totalRows int, err error)
	GetCertNewestValidOrderById(id int) (order Order, err error)

	PostNewOrder(payload NewOrderAcmePayload) (newId int, err error)

	PutOrderAcme(payload UpdateAcmeOrderPayload) (err error)
	PutOrderInvalid(orderId int) (err error)
	UpdateFinalizedKey(orderId int, keyId int) (err error)
	UpdateOrderCert(orderId int, CertPayload CertPayload) (err error)
	RevokeOrder(orderId int) (err error)

	GetAllValidCurrentOrders(q pagination_sort.Query) (orders []Order, totalRows int, err error)
	GetAllIncompleteOrderIds() (orderIds []int, err error)
	GetExpiringCertIds(maxTimeRemaining time.Duration) (certIds []int, err error)
	GetNewestIncompleteCertOrderId(certId int) (orderId int, err error)

	// certs
	UpdateCertUpdatedTime(certId int) (err error)
}

// Configuration options
type Config struct {
	AutomaticOrderingEnable     *bool `yaml:"auto_order_enable"`
	ValidRemainingDaysThreshold *int  `yaml:"valid_remaining_days_threshold"`
	RefreshTimeHour             *int  `yaml:"refresh_time_hour"`
	RefreshTimeMinute           *int  `yaml:"refresh_time_minute"`
}

// service struct
type Service struct {
	shutdownContext   context.Context
	logger            *zap.SugaredLogger
	output            *output.Service
	storage           Storage
	acmeServerService *acme_servers.Service
	authorizations    *authorizations.Service
	certificates      *certificates.Service

	serverCertificateName    *string
	loadHttpsCertificateFunc func() error
	httpClient               *httpclient.Client
	shellPath                string

	postProcessing  *job_manager.Manager[*postProcessJob]
	orderFulfilling *job_manager.Manager[*orderFulfillJob]
}

// NewService creates a new private_key service
func NewService(app App, cfg *Config) (*Service, error) {
	service := new(Service)

	// shutdown context
	service.shutdownContext = app.GetShutdownContext()

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// output service
	service.output = app.GetOutputter()
	if service.output == nil {
		return nil, errServiceComponent
	}

	// storage
	service.storage = app.GetOrderStorage()
	if service.storage == nil {
		return nil, errServiceComponent
	}

	// acme services
	service.acmeServerService = app.GetAcmeServerService()
	if service.acmeServerService == nil {
		return nil, errServiceComponent
	}

	// auths
	service.authorizations = app.GetAuthsService()
	if service.authorizations == nil {
		return nil, errServiceComponent
	}

	// certificates
	service.certificates = app.GetCertificatesService()
	if service.certificates == nil {
		return nil, errServiceComponent
	}

	// needed to reload App cert on update
	service.serverCertificateName = app.HttpsCertificateName()
	service.loadHttpsCertificateFunc = app.LoadHttpsCertificate

	// 	shell path for post processing
	var err error
	service.shellPath, err = exec.LookPath("powershell.exe")
	if err != nil {
		service.logger.Debugf("orders: unable to find powershell (%s)", err)
		// then try bash
		service.shellPath, err = exec.LookPath("bash")
		if err != nil {
			service.logger.Debugf("orders: unable to find bash (%s)", err)
			// then try zshell
			service.shellPath, err = exec.LookPath("zsh")
			if err != nil {
				logger.Debugf("orders: unable to find zshell (%s)", err)
				// then try sh
				service.shellPath, err = exec.LookPath("sh")
				if err != nil {
					logger.Debugf("orders: unable to find sh (%s)", err)
					// failed - disable post processing
					logger.Errorf("orders: unable to find a suitable shell for certificate post processing scripts (scripts disabled)")
					service.shellPath = ""
				}
			}
		}
	}

	// httpClient
	service.httpClient = app.GetHttpClient()

	// make post process job manager
	postWorkers := 3
	service.postProcessing = job_manager.NewManager[*postProcessJob](postWorkers, "post processing", app.GetShutdownContext(), app.GetShutdownWaitGroup(), app.GetLogger())
	if service.postProcessing == nil {
		return nil, errServiceComponent
	}

	// make order fulfill job manager
	fulfillingWorkers := 3
	service.orderFulfilling = job_manager.NewManager[*orderFulfillJob](fulfillingWorkers, "order fulfilling", app.GetShutdownContext(), app.GetShutdownWaitGroup(), app.GetLogger())
	if service.orderFulfilling == nil {
		return nil, errServiceComponent
	}

	// start service to automatically place and complete orders
	service.startAutoOrderService(cfg, app.GetShutdownContext(), app.GetShutdownWaitGroup())

	return service, nil
}
