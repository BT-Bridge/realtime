package main

import (
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/realtime"
	"gitlab.bcc-hyperdev.org/bcc-hyperdev/realtime/shared"
)

// OPENAI_API_KEY, OPENAI_ORG_ID, OPENAI_PROJECT_ID, OPENAI_WEBHOOK_SECRET, OPENAI_BASE_URL

var (
	apiKey    = shared.MustGetenv(shared.GetenvString, "OPENAI_API_KEY", true)
	orgId     = shared.MustGetenv(shared.GetenvString, "OPENAI_ORG_ID", true)
	projectId = shared.MustGetenv(shared.GetenvString, "OPENAI_PROJECT_ID", true)
)

func main() {
	client := realtime.NewRealtimeService(
		option.WithAPIKey(apiKey),
		option.WithOrganization(orgId),
		option.WithProject(projectId),
	)
	_ = client
}
