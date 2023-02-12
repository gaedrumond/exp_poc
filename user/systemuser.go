package systemuser

import (
	"os/user"
	"strings"

	"github.com/prometheus-community/windows_exporter/log"
)

func GetCurrentUser() *user.User {
	u, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	return u
}

func ValidateUser(user *user.User) {
	if strings.Contains(user.Username, "ContainerAdministrator") || strings.Contains(user.Username, "ContainerUser") {
		log.Warnf("Running as a preconfigured Windows Container user. This may mean you do not have Windows HostProcess containers configured correctly and some functionality will not work as expected.")
	}
}
