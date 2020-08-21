package main

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/prometheus/documentation/examples/custom-sd/adapter"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"prometheus-nacos-sd/nacos"
)

var (
	Version       string = "1.0"
	kp                   = kingpin.New("nacos sd adapter usage", "Tool to generate file_sd target files for implemented nacos SD mechanisms.")
	outputFile           = kp.Flag("output.file", "Output file for file_sd compatible file.").Default("nacos_sd.json").String()
	listenAddress        = kp.Flag("nacos.address", "The address Nacos listening on for requests.").Default("localhost:8848").String()
	namespace            = kp.Flag("nacos.namespace", "nacos public").Default("public").String()
	group                = kp.Flag("nacos.group", "nacos group").Default("DEFAULT_GROUP").String()
	refreshInterval      = kp.Flag("refresh.interval", "generate interval").Default("60").Int()
	logger        log.Logger
)

func main() {
	kp.Version(Version)
	kp.HelpFlag.Short('h')

	_, err := kp.Parse(os.Args[1:])
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	logger = log.NewSyncLogger(log.NewLogfmtLogger(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	// NOTE: create an instance of your new SD implementation here.
	cfg := nacos.NacosDiscovery{
		Address:         *listenAddress,
		Namespace:       *namespace,
		Group:           *group,
		TagSeparator:    ",",
		RefreshInterval: *refreshInterval,
		Logger:          logger,
	}

	disc, err := nacos.NewDiscovery(cfg)
	if err != nil {
		fmt.Println("err: ", err)
	}

	//create nacos naming client , put client into ctx
	nacosNamingClient := nacos.GetNacosNamingClient(cfg)
	ctx := context.WithValue(context.Background(), "nacosClient", nacosNamingClient)

	sdAdapter := adapter.NewAdapter(ctx, *outputFile, "nacosSD", disc, logger)
	sdAdapter.Run()

	<-ctx.Done()
}
