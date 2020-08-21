module prometheus-nacos-sd

go 1.13

require (
	github.com/Azure/azure-sdk-for-go v44.0.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.2 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.0 // indirect
	github.com/go-kit/kit v0.10.0
	github.com/hashicorp/consul/api v1.5.0 // indirect
	github.com/hashicorp/go-hclog v0.12.2 // indirect
	github.com/nacos-group/nacos-sdk-go v1.0.0
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.10.0
	github.com/prometheus/prometheus v1.8.2-0.20200805170718-983ebb4a5133
	github.com/prometheus/tsdb v0.10.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.18.5 // indirect
	k8s.io/apimachinery v0.18.5 // indirect
	k8s.io/client-go v0.18.5 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
)

replace k8s.io/klog => github.com/simonpasquier/klog-gokit v0.1.0
