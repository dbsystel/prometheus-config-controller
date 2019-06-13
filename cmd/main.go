package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/dbsystel/kube-controller-dbsystel-go-common/controller/configmap"
	"github.com/dbsystel/kube-controller-dbsystel-go-common/kubernetes"
	k8sflag "github.com/dbsystel/kube-controller-dbsystel-go-common/kubernetes/flag"
	opslog "github.com/dbsystel/kube-controller-dbsystel-go-common/log"
	logflag "github.com/dbsystel/kube-controller-dbsystel-go-common/log/flag"
	"github.com/dbsystel/prometheus-config-controller/controller"
	"github.com/dbsystel/prometheus-config-controller/prometheus"
	"github.com/go-kit/kit/log/level"
)

var (
	app = kingpin.New(filepath.Base(os.Args[0]), "Prometheus Controller")
	//Here you can define more flags for your application
	configPath     = app.Flag("config-path", "The location to save rule and config files to").Required().String()
	configTemplate = app.Flag("config-template", "The template of prometheus.yml").Required().String()
	id             = app.Flag("id", "The id of Prometheus").Default("0").Int()
	key            = app.Flag("key", "The unique key for prometheus config").String()
	reloadUrl      = app.Flag("reload-url", "The url to issue requests to reload Prometheus to").Required().String()
)

func main() {
	//Define config for logging
	var logcfg opslog.Config
	//Definie if controller runs outside of k8s
	var runOutsideCluster bool
	//Add two additional flags to application for logging and decision if inside or outside k8s
	logflag.AddFlags(app, &logcfg)
	k8sflag.AddFlags(app, &runOutsideCluster)
	//Parse all arguments
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		//Received error while parsing arguments from function app.Parse
		fmt.Fprintln(os.Stderr, "Catched the following error while parsing arguments: ", err)
		app.Usage(os.Args[1:])
		os.Exit(2)
	}
	//Initialize new logger from opslog
	logger, err := opslog.New(logcfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		app.Usage(os.Args[1:])
		os.Exit(2)
	}
	//First usage of initialized logger for testing
	level.Debug(logger).Log("msg", "Logging initiated...")
	//Initialize new k8s client from common k8s package
	k8sClient, err := kubernetes.NewClientSet(runOutsideCluster)
	if err != nil {
		level.Error(logger).Log("msg", err.Error())
		app.Usage(os.Args[1:])
		os.Exit(2)
	}

	rUrl, err := url.Parse(*reloadUrl)
	if err != nil {
		level.Error(logger).Log("msg", "Prometheus reload URL could not be parsed: "+*reloadUrl)
		os.Exit(2)
	}

	p := prometheus.New(rUrl, *configPath, *configTemplate, *id, *key, logger)

	level.Info(logger).Log("msg", "Starting Prometheus Controller...")
	sigs := make(chan os.Signal, 1) // Create channel to receive OS signals
	stop := make(chan struct{})     // Create channel to receive stop signal

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT) // Register the sigs channel to receieve SIGTERM

	wg := &sync.WaitGroup{} // Goroutines can add themselves to this to be waited on so that they finish

	//Initialize new k8s configmap-controller from common k8s package
	configMapController := &configmap.ConfigMapController{}
	configMapController.Controller = controller.New(*p, logger)
	configMapController.Initialize(k8sClient)
	//Run initiated configmap-controller as go routine
	go configMapController.Run(stop, wg)

	<-sigs // Wait for signals (this hangs until a signal arrives)

	level.Info(logger).Log("msg", "Shutting down...")

	close(stop) // Tell goroutines to stop themselves
	wg.Wait()   // Wait for all to be stopped
}
