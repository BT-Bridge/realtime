package main

import (
	"fmt"

	"gitlab.bcc-hyperdev.org/bcc-hyperdev/realtime/shared"
	"go.uber.org/zap"
)

type RealtimeService interface {
	NewClient() (RealtimeClient, error)
}

type RealtimeClient interface {
	Logger() *shared.Logger
	State() ClientState
}

type ImplConfig interface {
	Name() string
}

type ClientState int

const (
	ClientStateInitial ClientState = iota
)

type baseClient struct {
	logger *shared.Logger
	state  ClientState
}

func newBaseClient(logger *shared.Logger) *baseClient {
	return &baseClient{
		logger: logger,
		state:  ClientStateInitial,
	}
}

func (c *baseClient) Logger() *shared.Logger {
	return c.logger
}

func (c *baseClient) State() ClientState {
	return c.state
}

type OpenaiRealtimeClient struct {
	*baseClient
}

func NewOpenaiRealtimeClient(logger *shared.Logger, apiKey, orgId, projectId, baseUrl string) (*OpenaiRealtimeClient, error) {
	return &OpenaiRealtimeClient{
		baseClient: newBaseClient(logger),
	}, nil
}

func NewRealtimeService(logger *shared.Logger, implConfig ImplConfig) (svc RealtimeService, err error) {
	switch config := implConfig.(type) {
	case *OpenaiImplConfig:
		svc, err = NewOpenaiRealtimeService(logger, config)
	default:
		svc, err = nil, fmt.Errorf("unknown implementation: %s", implConfig.Name())
	}
	if err != nil {
		return nil, err
	}
	return svc, nil
}

type OpenaiImplConfig struct {
	ApiKey    string
	OrgId     string
	ProjectId string
	BaseUrl   string
}

func (c *OpenaiImplConfig) Name() string {
	return "openai"
}

type OpenaiRealtimeService struct {
	logger *shared.Logger
	cfg    *OpenaiImplConfig
}

func NewOpenaiRealtimeService(logger *shared.Logger, cfg *OpenaiImplConfig) (s *OpenaiRealtimeService, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create OpenAI Realtime Service: %w", err)
		}
	}()
	return &OpenaiRealtimeService{
		logger: logger,
		cfg:    cfg,
	}, nil
}

func (s *OpenaiRealtimeService) NewClient() (c RealtimeClient, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create client: %w", err)
		}
	}()
	return NewOpenaiRealtimeClient(
		s.logger,
		s.cfg.ApiKey,
		s.cfg.OrgId,
		s.cfg.ProjectId,
		s.cfg.BaseUrl,
	)
}

func main() {
	logger := shared.NewLogger(
		zap.String("package", "realtime"),
		zap.String("example", "openai"),
	)
	svc, err := NewRealtimeService(
		logger,
		&OpenaiImplConfig{
			ApiKey:    shared.MustGetenv(shared.GetenvString, "OPENAI_API_KEY", true),
			OrgId:     shared.MustGetenv(shared.GetenvString, "OPENAI_ORG_ID", true),
			ProjectId: shared.MustGetenv(shared.GetenvString, "OPENAI_PROJECT_ID", true),
			BaseUrl:   shared.MustGetenv(shared.GetenvString, "OPENAI_BASE_URL", false, "https://api.openai.com/v1"),
		},
	)
	if err != nil {
		logger.NoCtxFatal(err.Error())
	}
	client, err := svc.NewClient()
	if err != nil {
		logger.NoCtxFatal(err.Error())
	}
	_ = client
}
