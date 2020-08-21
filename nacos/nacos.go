package nacos

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	nacosModel "github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/util/strutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	namespaceLabel   = model.MetaLabelPrefix + "nacos_namespace"
	groupLabel       = model.MetaLabelPrefix + "nacos_group"
	metricsPathLabel = model.MetricsPathLabel
	jobLabel         = model.JobLabel
	//instanceLabel       = model.InstanceLabel
)

// Note: create a config struct for your custom SD type here.
type NacosSDConfig struct {
	Address         string
	NameSpace       string
	Group           string
	Username        string
	Password        string
	TagSeparator    string
	RefreshInterval int
}

// Note: This is the struct with your implementation of the Discoverer interface (see Run function).
// Discovery retrieves target information from a Consul server and updates them via watches.
type NacosDiscovery struct {
	Address         string
	Namespace       string
	Group           string
	Username        string
	Password        string
	RefreshInterval int
	TagSeparator    string
	Logger          log.Logger
	OldSourceList   map[string]bool
}

func (d *NacosDiscovery) parseServiceInstance(Service nacosModel.Service, serviceName string, namespace string, group string) (*targetgroup.Group, error) {

	if len(Service.Hosts) == 0 {
		return nil, errors.New("no instance found")
	}
	var instances []nacosModel.Instance
	instances = Service.Hosts

	tgroup := targetgroup.Group{
		Source: serviceName,
		Labels: make(model.LabelSet),
	}

	tgroup.Targets = make([]model.LabelSet, 0, len(instances))

	for _, instance := range instances {

		// If the service address is not empty it should be used instead of the node address
		// since the service may be registered remotely through a different node.
		var addr string
		if instance.Ip != "" {
			addr = net.JoinHostPort(instance.Ip, fmt.Sprintf("%d", instance.Port))
		} else {
			addr = net.JoinHostPort(instance.Ip, fmt.Sprintf("%d", instance.Port))
		}

		//fetch real metric path into label: __metrics_path__
		//for spring boot2.x, if config server.servlet.context-path="/xxx"
		//real prometheus metric path is "/xxx/actuator/prometheus"
		realMetraicsPath, ok := instance.Metadata["context_path"]
		if !ok || realMetraicsPath == "/" {
			realMetraicsPath = "/actuator/prometheus"
		} else {
			realMetraicsPath += "/actuator/prometheus"
		}

		target := model.LabelSet{model.AddressLabel: model.LabelValue(addr)}
		labels := model.LabelSet{
			model.LabelName(namespaceLabel):   model.LabelValue(namespace),
			model.LabelName(groupLabel):       model.LabelValue(group),
			model.LabelName(metricsPathLabel): model.LabelValue(realMetraicsPath),
			model.LabelName(jobLabel):         model.LabelValue(instance.ServiceName),
		}
		tgroup.Labels = labels

		// Add all key/value pairs from the node's metadata as their own labels.
		for k, v := range instance.Metadata {
			name := strutil.SanitizeLabelName(k)
			tgroup.Labels[model.LabelName(model.MetaLabelPrefix+name)] = model.LabelValue(v)
		}

		tgroup.Targets = append(tgroup.Targets, target)
	}
	return &tgroup, nil
}

// Note: you must implement this function for your discovery implementation as part of the
// Discoverer interface. Here you should query your SD for it's list of known targets, determine
// which of those targets you care about (for example, which of Consuls known services do you want
// to scrape for metrics), and then send those targets as a target.TargetGroup to the ch channel.
func (d *NacosDiscovery) Run(ctx context.Context, ch chan<- []*targetgroup.Group) {
	for c := time.Tick(time.Duration(d.RefreshInterval) * time.Second); ; {

		nacosNamingClient := ctx.Value("nacosClient").(naming_client.INamingClient)
		serviceInfos, err := nacosNamingClient.GetAllServicesInfo(vo.GetAllServiceInfoParam{
			NameSpace: d.Namespace,
			GroupName: d.Group,
			PageNo:    uint32(1),
			PageSize:  uint32(1000),
		})

		if err != nil {
			level.Error(d.Logger).Log("msg", "Error get nacos services list", "err", err)
			time.Sleep(time.Duration(d.RefreshInterval) * time.Second)
			continue
		}
		if serviceInfos.Count == 0 {
			level.Error(d.Logger).Log("msg", "Error got blank nacos services list")
			time.Sleep(time.Duration(d.RefreshInterval) * time.Second)
			continue

		}

		srvs := serviceInfos.Doms
		var tgs []*targetgroup.Group
		// Note that we treat errors when querying specific consul services as fatal for this
		// iteration of the time.Tick loop. It's better to have some stale targets than an incomplete
		// list of targets simply because there may have been a timeout. If the service is actually
		// gone as far as consul is concerned, that will be picked up during the next iteration of
		// the outer loop.

		newSourceList := make(map[string]bool)
		for _, servicename := range srvs {

			Service, err := nacosNamingClient.GetService(vo.GetServiceParam{
				ServiceName: servicename,
				GroupName:   d.Group,
			})

			if err != nil {
				level.Error(d.Logger).Log("msg", "Error getting services info from nacos", "service", servicename, "err", err)
				break
			}
			tg, err := d.parseServiceInstance(Service, servicename, d.Namespace, d.Group)
			if err != nil {
				level.Error(d.Logger).Log("msg", "Error parsing services instance", "service", servicename, "err", err)
				break
			}
			tgs = append(tgs, tg)
			newSourceList[tg.Source] = true
		}
		// When targetGroup disappear, send an update with empty targetList.
		for key := range d.OldSourceList {
			if !newSourceList[key] {
				tgs = append(tgs, &targetgroup.Group{
					Source: key,
				})
			}
		}
		d.OldSourceList = newSourceList
		if err == nil {
			level.Info(d.Logger).Log("msg", "fetch services config done")
			ch <- tgs
		}
		// Wait for ticker or exit when ctx is closed.
		select {
		case <-c:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func NewDiscovery(conf NacosDiscovery) (*NacosDiscovery, error) {
	cd := &NacosDiscovery{
		Address:         conf.Address,
		Namespace:       conf.Namespace,
		Group:           conf.Group,
		RefreshInterval: conf.RefreshInterval,
		TagSeparator:    conf.TagSeparator,
		Logger:          conf.Logger,
		OldSourceList:   make(map[string]bool),
	}
	return cd, nil
}

func GetNacosNamingClient(conf NacosDiscovery) naming_client.INamingClient {
	var once sync.Once
	var nacosNamingClient naming_client.INamingClient
	once.Do(func() {
		nacosNamingClient = createNacosNamingClient(conf)
	})
	return nacosNamingClient
}

func createNacosNamingClient(conf NacosDiscovery) naming_client.INamingClient {

	address := strings.Split(conf.Address, ":")
	port, _ := strconv.ParseUint(address[1], 10, 64)
	var serverConfigs []constant.ServerConfig
	serverConfig := constant.ServerConfig{
		ContextPath: "/nacos",
		IpAddr:      address[0],
		Port:        port,
	}

	currentProcessPath, _ := os.Executable()
	cacheDir := os.TempDir() + string(os.PathSeparator) + filepath.Base(currentProcessPath) + string(os.PathSeparator) + "cache" + string(os.PathSeparator) + conf.Namespace
	logDir := os.TempDir() + string(os.PathSeparator) + filepath.Base(currentProcessPath) + string(os.PathSeparator) + "log" + string(os.PathSeparator) + conf.Namespace

	clientConfig := constant.ClientConfig{
		NamespaceId:          conf.Namespace,
		Username:             conf.Username,
		Password:             conf.Password,
		TimeoutMs:            uint64(3000),
		NotLoadCacheAtStart:  true,
		UpdateCacheWhenEmpty: true,
		CacheDir:             cacheDir,
		LogDir:               logDir,
		LogLevel:             "debug",
	}
	nacosNamingClient, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": append(serverConfigs, serverConfig),
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return nil
	}
	return nacosNamingClient
}
