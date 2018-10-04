package provisioner

type instanceProvisionerFactory struct {
	dind InstanceProvisionerApi
}

func NewInstanceProvisionerFactory(d InstanceProvisionerApi) InstanceProvisionerFactoryApi {
	return &instanceProvisionerFactory{dind: d}
}

func (p *instanceProvisionerFactory) GetProvisioner(instanceType string) (InstanceProvisionerApi, error) {
	return p.dind, nil
}
