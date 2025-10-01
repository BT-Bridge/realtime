package main

import (
	"fmt"

	"gitlab.bcc-hyperdev.org/bcc-hyperdev/realtime/shared"
	"go.uber.org/zap"
)

type OpenaiRealtimeClient struct {
	logger    *shared.Logger
	apiKey    string
	orgId     string
	projectId string
	baseUrl   string
}

func NewOpenaiRealtimeClient(logger *shared.Logger, apiKey, orgId, projectId, baseUrl string) (*OpenaiRealtimeClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey is required")
	}
	if orgId == "" {
		return nil, fmt.Errorf("orgId is required")
	}
	if projectId == "" {
		return nil, fmt.Errorf("projectId is required")
	}
	if baseUrl == "" {
		baseUrl = "https://api.openai.com/v1"
	}
	return &OpenaiRealtimeClient{
		logger:    logger,
		apiKey:    apiKey,
		orgId:     orgId,
		projectId: projectId,
		baseUrl:   baseUrl,
	}, nil
}

type OpenaiConfig struct {
	ApiKey    string
	OrgId     string
	ProjectId string
	BaseUrl   string
}

type OpenaiRealtimeService struct {
	logger *shared.Logger
	cfg    *OpenaiConfig
}

func NewOpenaiRealtimeService(logger *shared.Logger, cfg *OpenaiConfig) (s *OpenaiRealtimeService, err error) {
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

func (s *OpenaiRealtimeService) NewClient() (c *OpenaiRealtimeClient, err error) {
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
	svc, err := NewOpenaiRealtimeService(
		logger,
		&OpenaiConfig{
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
