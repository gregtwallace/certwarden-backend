package orders

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/domain/acme_servers"
	"legocerthub-backend/pkg/domain/authorizations"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary orders service component is missing")

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
	HttpsCertificateName() *string
	LoadHttpsCertificate() error
}

// Storage interface for storage functions
type Storage interface {
	// orders
	GetOneOrder(orderId int) (order Order, err error)
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

// Keys service struct
type Service struct {
	shutdownContext   context.Context
	logger            *zap.SugaredLogger
	output            *output.Service
	storage           Storage
	acmeServerService *acme_servers.Service
	certificates      *certificates.Service
	orderFulfiller    *orderFulfiller
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

	// certificates
	service.certificates = app.GetCertificatesService()
	if service.certificates == nil {
		return nil, errServiceComponent
	}

	// make worker manager
	workerCount := 3
	service.orderFulfiller = createOrderFulfiller(app, workerCount)

	// start service to automatically place and complete orders
	service.startAutoOrderService(cfg, service.shutdownContext, app.GetShutdownWaitGroup())

	return service, nil
}
