package saas

import (
	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerDetector(&WindowsUpdateDetector{})
}

type WindowsUpdateDetector struct{}

func (d *WindowsUpdateDetector) Name() string {
	return "Windows Update"
}

func (d *WindowsUpdateDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	if matchDNSSuffix(endpoint, []string{
		".windowsupdate.microsoft.com",
		".update.microsoft.com",
		".windowsupdate.com",
		"download.windowsupdate.com",
		"download.microsoft.com",
		".download.windowsupdate.com",
		"wustat.windows.com",
		"ntservicepack.microsoft.com",
		".delivery.mp.microsoft.com",
		"dl.delivery.mp.microsoft.com",
	}) {
		return true, nil
	}
	return false, nil
}
