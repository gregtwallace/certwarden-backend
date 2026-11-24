package orders

import (
	"certwarden-backend/pkg/datatypes/job_manager"
	"certwarden-backend/pkg/output"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"go.uber.org/zap"
)

type idleWorkStatusStorage struct {
	Storage
}

func (idleWorkStatusStorage) GetOrders(_ []int) ([]*Order, error) {
	return nil, sql.ErrNoRows
}

type workStatusOutputApp struct {
	logger *zap.SugaredLogger
}

func (app workStatusOutputApp) GetLogger() *zap.SugaredLogger {
	return app.logger
}

func TestIdleWorkStatusHandlersReturnEmptyStatus(t *testing.T) {
	service := newIdleWorkStatusTestService(t)

	tests := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request) *output.JsonError
	}{
		{
			name:    "fulfilling",
			handler: service.GetFulfillWorkStatus,
		},
		{
			name:    "post processing",
			handler: service.GetPostProcessWorkStatus,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/status", http.NoBody)

			if errJSON := test.handler(response, request); errJSON != nil {
				t.Fatalf("idle status returned an error: %v", errJSON)
			}
			if response.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
			}

			var body orderWorkStatusResponse
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if body.Message != "ok" {
				t.Errorf("expected message %q, got %q", "ok", body.Message)
			}
			if len(body.JobsWaiting) != 0 {
				t.Errorf("expected no waiting jobs, got %d", len(body.JobsWaiting))
			}
			for workerID, job := range body.JobsWorking {
				if job != nil {
					t.Errorf("expected worker %d to be idle", workerID)
				}
			}
		})
	}
}

func newIdleWorkStatusTestService(t *testing.T) *Service {
	t.Helper()

	logger := zap.NewNop().Sugar()
	outputService, err := output.NewService(workStatusOutputApp{logger: logger})
	if err != nil {
		t.Fatalf("failed to create output service: %v", err)
	}

	shutdownContext, cancel := context.WithCancel(context.Background())
	shutdownWaitGroup := new(sync.WaitGroup)
	t.Cleanup(func() {
		cancel()
		shutdownWaitGroup.Wait()
	})

	return &Service{
		logger:          logger,
		output:          outputService,
		storage:         idleWorkStatusStorage{},
		postProcessing:  job_manager.NewManager[*postProcessJob](1, "test post processing", shutdownContext, shutdownWaitGroup, logger),
		orderFulfilling: job_manager.NewManager[*orderFulfillJob](1, "test order fulfilling", shutdownContext, shutdownWaitGroup, logger),
		shutdownContext: shutdownContext,
	}
}
