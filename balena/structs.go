package balena

import (
	"github.com/google/uuid"
	"time"
)

type Fleet struct {
	id                       int
	actor                    int
	organization             int
	slug                     string
	appName                  string
	configuredRelease        int
	applicationType          int
	deviceType               int
	trackLatestRelease       bool
	accessibleBySupportUntil time.Time
	isPublic                 bool
	isHost                   bool
	isArchived               bool
	isDiscoverable           bool
	repositoryUrl            string
	createdAt                time.Time
	uuid                     uuid.UUID
}
