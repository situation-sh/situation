package saas

import (
	"net"

	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	// run curl -X GET "https://ip-ranges.datadoghq.com/" -H "Accept: application/json"
	// or look at https://docs.datadoghq.com/fr/api/latest/ip-ranges/?code-lang=curl
	registerDetector(&DatadogDetector{
		nets: []*net.IPNet{
			mustParseCIDR("100.28.212.0/22"),
			mustParseCIDR("107.21.25.247/32"),
			mustParseCIDR("108.137.133.223/32"),
			mustParseCIDR("108.137.188.57/32"),
			mustParseCIDR("13.114.211.96/32"),
			mustParseCIDR("13.115.46.213/32"),
			mustParseCIDR("13.126.169.175/32"),
			mustParseCIDR("13.208.126.217/32"),
			mustParseCIDR("13.208.133.55/32"),
			mustParseCIDR("13.208.142.17/32"),
			mustParseCIDR("13.208.255.200/32"),
			mustParseCIDR("13.209.118.42/32"),
			mustParseCIDR("13.209.230.111/32"),
			mustParseCIDR("13.234.54.8/32"),
			mustParseCIDR("13.236.246.161/32"),
			mustParseCIDR("13.238.14.57/32"),
			mustParseCIDR("13.244.188.203/32"),
			mustParseCIDR("13.244.85.86/32"),
			mustParseCIDR("13.245.194.43/32"),
			mustParseCIDR("13.245.200.254/32"),
			mustParseCIDR("13.246.172.210/32"),
			mustParseCIDR("13.247.164.9/32"),
			mustParseCIDR("13.48.150.244/32"),
			mustParseCIDR("13.48.239.118/32"),
			mustParseCIDR("13.48.254.37/32"),
			mustParseCIDR("13.54.169.48/32"),
			mustParseCIDR("15.152.238.192/32"),
			mustParseCIDR("15.161.86.71/32"),
			mustParseCIDR("15.165.240.116/32"),
			mustParseCIDR("15.168.188.85/32"),
			mustParseCIDR("15.184.139.182/32"),
			mustParseCIDR("15.185.189.82/32"),
			mustParseCIDR("15.188.202.64/32"),
			mustParseCIDR("15.188.240.172/32"),
			mustParseCIDR("15.188.243.248/32"),
			mustParseCIDR("157.241.36.106/32"),
			mustParseCIDR("157.241.93.102/32"),
			mustParseCIDR("16.162.136.62/32"),
			mustParseCIDR("16.163.153.45/32"),
			mustParseCIDR("16.24.38.13/32"),
			mustParseCIDR("16.24.60.114/32"),
			mustParseCIDR("18.102.80.189/32"),
			mustParseCIDR("18.130.113.168/32"),
			mustParseCIDR("18.139.52.173/32"),
			mustParseCIDR("18.163.21.55/32"),
			mustParseCIDR("18.163.59.106/32"),
			mustParseCIDR("18.166.19.255/32"),
			mustParseCIDR("18.195.155.52/32"),
			mustParseCIDR("18.200.120.237/32"),
			mustParseCIDR("18.229.28.50/32"),
			mustParseCIDR("18.229.36.120/32"),
			mustParseCIDR("20.62.248.141/32"),
			mustParseCIDR("20.83.144.189/32"),
			mustParseCIDR("23.20.198.65/32"),
			mustParseCIDR("23.23.216.60/32"),
			mustParseCIDR("3.120.223.25/32"),
			mustParseCIDR("3.121.24.234/32"),
			mustParseCIDR("3.1.219.207/32"),
			mustParseCIDR("3.1.36.99/32"),
			mustParseCIDR("3.18.172.189/32"),
			mustParseCIDR("3.18.188.104/32"),
			mustParseCIDR("3.18.197.0/32"),
			mustParseCIDR("3.210.147.169/32"),
			mustParseCIDR("3.220.254.141/32"),
			mustParseCIDR("3.233.144.0/20"),
			mustParseCIDR("3.35.66.96/32"),
			mustParseCIDR("3.36.177.119/32"),
			mustParseCIDR("34.145.82.128/29"),
			mustParseCIDR("34.146.154.144/29"),
			mustParseCIDR("34.159.50.128/29"),
			mustParseCIDR("34.174.98.16/29"),
			mustParseCIDR("34.203.1.9/32"),
			mustParseCIDR("34.204.83.4/32"),
			mustParseCIDR("34.208.32.189/32"),
			mustParseCIDR("34.233.140.66/32"),
			mustParseCIDR("34.48.76.208/29"),
			mustParseCIDR("34.94.234.88/29"),
			mustParseCIDR("35.152.76.8/32"),
			mustParseCIDR("35.154.93.182/32"),
			mustParseCIDR("35.172.176.208/32"),
			mustParseCIDR("35.176.195.46/32"),
			mustParseCIDR("35.177.43.250/32"),
			mustParseCIDR("3.92.150.182/32"),
			mustParseCIDR("3.96.7.126/32"),
			mustParseCIDR("40.76.107.170/32"),
			mustParseCIDR("43.198.123.228/32"),
			mustParseCIDR("43.203.72.233/32"),
			mustParseCIDR("43.218.5.202/32"),
			mustParseCIDR("44.192.28.0/25"),
			mustParseCIDR("52.1.33.14/32"),
			mustParseCIDR("52.1.61.69/32"),
			mustParseCIDR("52.192.175.207/32"),
			mustParseCIDR("52.35.61.232/32"),
			mustParseCIDR("52.55.56.26/32"),
			mustParseCIDR("52.60.189.53/32"),
			mustParseCIDR("52.67.95.251/32"),
			mustParseCIDR("52.89.221.151/32"),
			mustParseCIDR("52.9.13.199/32"),
			mustParseCIDR("52.9.139.134/32"),
			mustParseCIDR("54.157.36.5/32"),
			mustParseCIDR("54.177.155.33/32"),
			mustParseCIDR("54.92.248.81/32"),
			mustParseCIDR("63.34.100.178/32"),
			mustParseCIDR("63.35.33.198/32"),
			mustParseCIDR("99.79.87.237/32"),
			mustParseCIDR("2600:1f18:24e6:b900::/56"),
		},
	})
}

type DatadogDetector struct {
	nets []*net.IPNet
}

func (d *DatadogDetector) Name() string {
	return "Datadog"
}

func (d *DatadogDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	return matchIPRange(endpoint, d.nets)
}
