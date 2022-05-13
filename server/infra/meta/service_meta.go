package meta

type ServiceInfo struct {
	ServiceIdentity
	Build   *Build         `json:"build"`
	Runtime *RuntimeStatus `json:"runtime"`
	Brand   *Brand         `json:"brand"`
}

var AcquireServiceMetaFunc = AcquireServiceMeta

func AcquireServiceMeta() ServiceInfo {
	buildInfo, _ := AcquireBuildInfoFunc(buildInfoPath)
	runtimeInfo := AcquireRuntimeStatusFunc()
	brand, _ := AcquireBrandInfoFunc(brandPath)

	return ServiceInfo{
		ServiceIdentity: AcquireServiceIdentityFunc(),
		Build:           buildInfo,
		Runtime:         &runtimeInfo,
		Brand:           brand,
	}
}
