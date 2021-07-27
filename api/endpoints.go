package api

import (
	"net/http"

	"github.com/dbubel/intake"
)

func GetCredentialEndpoints(c Credentials, auth intake.MiddleWare) intake.Endpoints {
	return intake.Endpoints{
		intake.NewEndpoint(http.MethodPost, "/users/credentials", c.createCredential, auth),
		intake.NewEndpoint(http.MethodGet, "/users/credentials", c.getCredentials, auth),
		intake.NewEndpoint(http.MethodGet, "/users/credentials/:credentialUid", c.getCredential, auth),
		intake.NewEndpoint(http.MethodPut, "/users/credentials/:credentialUid/username", c.updateUsername, auth),
		intake.NewEndpoint(http.MethodPut, "/users/credentials/:credentialUid/password", c.updatePassword, auth),
		intake.NewEndpoint(http.MethodPut, "/users/credentials/:credentialUid/service", c.updateServiceName, auth),
		intake.NewEndpoint(http.MethodDelete, "/users/credentials/:credentialUid", c.deleteCredential, auth),
		intake.NewEndpoint(http.MethodGet, "/status", c.status),
	}
}
func GetUserManagementEndpoints(c FireBaseAuth) intake.Endpoints {
	return intake.Endpoints{
		intake.NewEndpoint(http.MethodPost, "/users/signin", c.Signin),
	}
}
