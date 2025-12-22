package init

import services_interface "github.com/Identityplane/GoAM/pkg/services"

type ServicesFactory interface {
	CreateServices() (*services_interface.Services, error)
}

var servicesFactory ServicesFactory

func SetServicesFactory(factory ServicesFactory) {
	servicesFactory = factory
}

func GetServicesFactory() ServicesFactory {
	return servicesFactory
}
