package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"runtime"
	"runtime/pprof"
	"github.com/leapar/libvirt-metrics/config"
	"github.com/leapar/libvirt-metrics/backend"
	"github.com/leapar/libvirt-metrics/libvirt"

	"github.com/takama/daemon"

	"os/signal"
	"syscall"
)

const (
	// name of the service
	name        = "libvirt-tsdb"
	description = "send kvm stats to opentsdb"
)

var dependencies = []string{}

var stdlog, errlog *log.Logger

// Service has embedded daemon
type Service struct {
	daemon.Daemon
}



func queryKvm(kvm libvirt.KVM, config config.Configuration, channel *chan []backend.Point) {
	kvm.Query(config.Interval,  channel)
}

func queryKvmHost(kvm libvirt.KVM, channel *chan backend.HostStuct) {
	kvm.QueryHost(channel)
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {
	//defer saveHeapProfile()
	usage := "Usage: myservice install | remove | start | stop | status"

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	stdlog.Println("Starting daemon:", path.Base(os.Args[0]))

	// read the configuration
	file, err := os.Open("/etc/" + path.Base(os.Args[0]) + ".json")
	if err != nil {
		return "Could not open configuration file", err
	}
	jsondec := json.NewDecoder(file)
	config := config.Configuration{}
	err = jsondec.Decode(&config)
	if err != nil {
		return "Could not decode configuration file", err
	}

	//force backend values to environement varialbles if present
	s := reflect.ValueOf(&config.Backend).Elem()
	numfields := s.NumField()
	for i := 0; i < numfields; i++ {
		f := s.Field(i)
		if f.CanSet() {
			//exported field
			envname := strings.ToUpper(s.Type().Name() + "_" + s.Type().Field(i).Name)
			envval := os.Getenv(envname)
			if len(envval) > 0 {
				//environment variable set with name
				switch ftype := f.Type().Name(); ftype {
				case "string":
					f.SetString(envval)
				case "int":
					val, err := strconv.ParseInt(envval, 10, 64)
					if err == nil {
						f.SetInt(val)
					}
				}
			}
		}
	}


	err = config.Backend.Init(stdlog, errlog)
	if err != nil {
		return "Could not initialize backend", err
	}
	defer config.Backend.Disconnect()

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Set up a channel to recieve the metrics
	metrics := make(chan []backend.Point)
	finders := make(chan backend.HostStuct)

	// Set up a ticker to collect metrics at givent interval
	ticker := time.NewTicker(time.Second * time.Duration(config.Interval))
	defer ticker.Stop()

	tickerFinder := time.NewTicker(time.Second * time.Duration(config.Interval) )//* 20
	defer tickerFinder.Stop()

	// Start retriveing and sending metrics
	stdlog.Println("Retrieving metrics")
	for _, kvm := range config.Kvms {
		go queryKvm(*kvm, config, &metrics)
		go queryKvmHost(*kvm, &finders)

	}
	for {
		select {
		case values := <-metrics:
			config.Backend.SendMetrics(values)
			stdlog.Printf("Sent %d metrics to backend", len(values))
		case values := <-finders:

			url := fmt.Sprintf("%s?api_key=%s&host=%s", config.Backend.HostUrl, config.Backend.ApiKey, values.Host)
			config.Backend.SendHost(values.Infos, url)
			stdlog.Printf("Sent %d host info to backend", len(values.Infos))
		case <-ticker.C:
			stdlog.Println("Retrieving hosts")
			for _, kvm := range config.Kvms {
				go queryKvm(*kvm, config, &metrics)
			}
		case <-tickerFinder.C:
			stdlog.Println("Retrieving metrics")
			for _, kvm := range config.Kvms {
				go queryKvmHost(*kvm, &finders)
			}
		case killSignal := <-interrupt:
			stdlog.Println("Got signal:", killSignal)
			if killSignal == os.Interrupt {
				return "Daemon was interruped by system signal", nil
			}
			return "Daemon was killed", nil
		}
	}

	// never happen, but need to complete code
	return usage, nil
}

var (
	pid      int
	progname string
)

func init() {
	pid = os.Getpid()
	paths := strings.Split(os.Args[0], "/")
	paths = strings.Split(paths[len(paths)-1], string(os.PathSeparator))
	progname = paths[len(paths)-1]
	runtime.MemProfileRate = 1

	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func saveHeapProfile() {
	runtime.GC()
	f, err := os.Create(fmt.Sprintf("heap_%s_%d_%s.prof", progname, pid, time.Now().Format("2006_01_02_03_04_05")))
	if err != nil {
		return
	}
	defer f.Close()
	pprof.Lookup("heap").WriteTo(f, 1)
	fmt.Println("saveHeapProfile")
}

func main() {
	srv, err := daemon.New(name, description, dependencies...)
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
