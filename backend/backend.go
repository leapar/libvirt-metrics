package backend

import (
	"log"
	"strings"
	"errors"
	"time"
	"strconv"
	"github.com/libvirt/libvirt-go"

	"github.com/olegfedoseev/opentsdb"
	"bytes"
	"github.com/pquerna/ffjson/ffjson"
	"compress/gzip"
	"net/http"
	"fmt"
)

type HostStuct struct {
	Host  string
	Infos []HostInfo
}

type HostInfo struct {
	Id   uint
	Name string
	Info *libvirt.DomainInfo
}

type Point struct {
	KvmHost string
	Domain  string
	Group   string
	Desc    string
	Metric  string
	Tags map[string]string
	Value      uint64
	Timestamp  int64
}

type KeyValue struct {
	PollerIp string
}

//Storage backend
type Backend struct {
	ApiKey    string
	MetricUrl string
	HostUrl   string
	Hostname  string
	Port      int
	Database  string
	Username  string
	Password  string
	Type      string
	KongHost  string
	PollerUrl string
	NoArray   bool
	opentsdb  *opentsdb.Client
	PollerTags KeyValue
}

var stdlog, errlog *log.Logger

func (backend *Backend) Init(standardLogs *log.Logger, errorLogs *log.Logger) error {
	stdlog = standardLogs
	errlog = errorLogs
	switch backendType := strings.ToLower(backend.Type); backendType {
	case "kong":
		return nil
	default:
		errlog.Println("Backend " + backendType + " unknown.")
		return errors.New("Backend " + backendType + " unknown.")
	}
}

func (backend *Backend) Disconnect() {

	switch backendType := strings.ToLower(backend.Type); backendType {

	case "opentsdb":
		stdlog.Println("Disconnecting from " + backendType)
	case "kong":

	default:
		errlog.Println("Backend " + backendType + " unknown.")
	}
}


func (backend *Backend) SendMetrics(metrics []Point) {
	switch backendType := strings.ToLower(backend.Type); backendType {



	case "kong":
		var tsdbMetrics opentsdb.DataPoints
		var host string
		for _, point := range metrics {
			//key := "libvirt." + vcName + "." + entityName + "." + name + "." + metricName
			/*key :=  "libvirt." + point.VCenter + "." + point.ObjectType + "." + point.ObjectName + "." + point.Group + "." + point.Counter + "." + point.Rollup
			if len(point.Instance) > 0 {
				key += "." + strings.ToLower(strings.Replace(point.Instance, ".", "_", -1))
			}*/



			tags := opentsdb.Tags{}
			if host == "" {
				host = point.KvmHost
			}



			for k, v := range point.Tags {
				tags[k] = v
			}



			tags["host"] = point.Domain
			tags["kvm-host"] = point.KvmHost


			tsdbMetrics = append(tsdbMetrics, &opentsdb.DataPoint{
				Metric:    point.Group + "." + point.Metric,
				Value:     strconv.FormatUint(point.Value, 10),
				Timestamp: time.Now().Unix(),
				Tags:      tags})
		}
		//b, _:= json.Marshal(tsdbMetrics)
		//fmt.Println(string(b))

		url := fmt.Sprintf("%s%s?api_key=%s&host=%s", backend.KongHost, backend.MetricUrl, backend.ApiKey, host)
		backend.SendNetrics2tsdb(tsdbMetrics, url)

		tsdbMetrics = nil
		//err := backend.carbon.SendMetrics(graphiteMetrics)
		//if err != nil {
		//	errlog.Println("Error sending metrics (trying to reconnect): ", err)
		//	backend.carbon.Connect()
		//}

	default:
		errlog.Println("Backend " + backendType + " unknown.")
	}
}

func (backend *Backend) SendNetrics2tsdb(values opentsdb.DataPoints, url string) (error) {
	var buffer bytes.Buffer

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
		},
	}
	writer := gzip.NewWriter(&buffer)

	if err := ffjson.NewEncoder(writer).Encode(values); err != nil {
		errlog.Println("send error:" + err.Error())
		return err
	}
	if err := writer.Close(); err != nil {
		errlog.Println("send error:" + err.Error())
		return err
	}

	req, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		errlog.Println("send error:" + err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		errlog.Println("send error:" + err.Error())
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (backend *Backend) SendHost(values []HostInfo, url string) (error) {
	var buffer bytes.Buffer

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
		},
	}
	writer := gzip.NewWriter(&buffer)

	if err := ffjson.NewEncoder(writer).Encode(values); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (backend *Backend) SendPollerMetrics(url string) (error) {
	var buffer bytes.Buffer

	var tsdbMetrics opentsdb.DataPoints
	tags := opentsdb.Tags{}

	tags["poller"] = backend.PollerTags.PollerIp;
	tags["type"] = "kvm";

	tsdbMetrics = append(tsdbMetrics, &opentsdb.DataPoint{
		Metric:    "poller.up",
		Value:     1,
		Timestamp: time.Now().Unix(),
		Tags:      tags})

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
		},
	}
	writer := gzip.NewWriter(&buffer)

	if err := ffjson.NewEncoder(writer).Encode(tsdbMetrics); err != nil {
		errlog.Println("send error:" + err.Error())
		return err
	}
	if err := writer.Close(); err != nil {
		errlog.Println("send error:" + err.Error())
		return err
	}

	req, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		errlog.Println("send error:" + err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		errlog.Println("send error:" + err.Error())
		return err
	}

	defer resp.Body.Close()

	return nil
}
