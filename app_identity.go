package framing

import "github.com/google/uuid"

type AppIdentity struct {
	Name        string
	Description string
	UUID        uuid.UUID
	AppDomain   string
}

func NewAppIdentity(appdomain string, name string, description string) AppIdentity {
	uuid := uuid.New()
	return AppIdentity{
		Name:        name,
		Description: name,
		AppDomain:   appdomain,
		UUID:        uuid,
	}
}
