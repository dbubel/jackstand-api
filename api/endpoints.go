package api

import (
	"net/http"

	"github.com/bf-dbubel/intake"
)

func endpoints(c Credentials) intake.Endpoints {
	return intake.Endpoints{
		intake.NewEndpoint(http.MethodPost, "/users/credentials", c.createCredential),
		intake.NewEndpoint(http.MethodGet, "/users/credentials", c.getCredentials),
		intake.NewEndpoint(http.MethodGet, "/users/credentials/:credentialUid", c.getCredential),
		intake.NewEndpoint(http.MethodPut, "/users/credentials/:credentialUid/username", c.updateUsername),
		intake.NewEndpoint(http.MethodPut, "/users/credentials/:credentialUid/password", c.updatePassword),
		intake.NewEndpoint(http.MethodPut, "/users/credentials/:credentialUid/service", c.updateServiceName),
		intake.NewEndpoint(http.MethodDelete, "/users/credentials/:credentialUid", c.deleteCredential),
	}
}
