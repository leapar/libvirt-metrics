package libvirt

import (
	"log"
	"github.com/leapar/libvirt-metrics/backend"
	"github.com/libvirt/libvirt-go"
	"fmt"
	"github.com/kumina/libvirt_exporter/libvirt_schema"
	"encoding/xml"
)

var stdlog, errlog *log.Logger

// VCenter description
//qemu+tcp://172.29.231.80/system
type KVM struct {
	Hostname     string
	MetricGroups []*MetricGroup
}

// Metric Definition
type MetricDef struct {
	Metric string
}

// Metric Grouping for retrieval
type MetricGroup struct {
	Metrics []MetricDef
}

// Metrics description in config
type Metric struct {
	Definition []MetricDef
}

func (vcenter *KVM) Connect() (*libvirt.Connect, error) {
	var virConn *libvirt.Connect
	var err error

	url := fmt.Sprintf("qemu+tcp://%s/system", vcenter.Hostname)
	virConn, err = libvirt.NewConnect(url)

	return virConn, err
}

// CollectDomain extracts Prometheus metrics from a libvirt domain.
func (kvm *KVM) CollectDomain(domain *libvirt.Domain) ([]backend.Point, error) {
	domainName, err := domain.GetName()
	if err != nil {
		return nil, err
	}

	// Decode XML description of domain to get block device names, etc.
	xmlDesc, err := domain.GetXMLDesc(0)
	if err != nil {
		return nil, err
	}
	var desc libvirt_schema.Domain
	err = xml.Unmarshal([]byte(xmlDesc), &desc)
	if err != nil {
		return nil, err
	}

	// Report domain info.
	info, err := domain.GetInfo()
	if err != nil {
		return nil, err
	}

	var points []backend.Point

	//dominfo                        域信息
	//maximum_memory_bytes
	//Maximum allowed memory of the domain, in bytes.
	/*
		virsh # dominfo ubuntu1
		Id:             1
		名称：       ubuntu1
		UUID:           0516216a-783a-4f1a-ab08-47efd4190cd8
		OS 类型：    hvm
		状态：       running
		CPU：          1
		CPU 时间：   18.1s
		最大内存： 1048576 KiB
		使用的内存： 1048576 KiB
		持久：       是
		自动启动： 禁用
		管理的保存： 否
		安全性模式： selinux
		安全性 DOI： 0
		安全性标签： system_u:system_r:svirt_t:s0:c580,c743 (enforcing)

	*/

	points = append(points, backend.Point{

		KvmHost: kvm.Hostname,
		Domain:  domainName,
		Group:   "domain.info",
		Desc:    "Maximum allowed memory of the domain, in bytes.",
		Metric:  "mem.max",
		Value:   info.MaxMem,
	})

	points = append(points, backend.Point{
		KvmHost: kvm.Hostname,
		Domain:  domainName,
		Group:   "domain.info",
		Desc:    "Memory usage of the domain, in bytes.",
		Metric:  "mem.usage",
		Value:   info.Memory,
	})

	points = append(points, backend.Point{
		KvmHost: kvm.Hostname,
		Domain:  domainName,
		Group:   "domain.info",
		Desc:    "Number of virtual CPUs for the domain.",
		Metric:  "cpu.num",
		Value:   uint64(info.NrVirtCpu),
	})

	points = append(points, backend.Point{
		KvmHost: kvm.Hostname,
		Domain:  domainName,
		Group:   "domain.info",
		Desc:    "Amount of CPU time used by the domain, in seconds.",
		Metric:  "cpu.time",
		Value:   info.CpuTime,
	})

	// Report block device statistics.
	for _, disk := range desc.Devices.Disks {
		blockStats, err := domain.BlockStats(disk.Target.Device)
		if err != nil {
			return points, err
		}

		/*

		domblkstat ubuntu1 hda
		hda rd_req 1
		hda rd_bytes 512
		hda wr_req 0
		hda wr_bytes 0
		hda flush_operations 0
		hda rd_total_times 2533293
		hda wr_total_times 0
		hda flush_total_times 0

		*/
		var tags map[string]string
		tags = map[string]string {}
		tags["File"] = disk.Source.File;
		tags["Device"] = disk.Target.Device;
		if blockStats.RdBytesSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.block.stats",
				Desc:    "Number of bytes read from a block device, in bytes.",
				Metric:  "rd.bytes",
				Value:   uint64(blockStats.RdBytes),
				Tags:    tags,
			})
		}

		if blockStats.RdReqSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.block.stats",
				Desc:    "Number of read requests from a block device.",
				Metric:  "rd.req",
				Value:   uint64(blockStats.RdReq),
				Tags:    tags,
			})
		}

		if blockStats.RdTotalTimesSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.block.stats",
				Desc:    "Amount of time spent reading from a block device, in seconds.",
				Metric:  "rd.total.times",
				Value:   uint64(blockStats.RdTotalTimes),
				Tags:    tags,
			})
		}

		if blockStats.WrBytesSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.block.stats",
				Desc:    "Number of bytes written from a block device, in bytes.",
				Metric:  "wr.bytes",
				Value:   uint64(blockStats.WrBytes),
				Tags:    tags,
			})
		}

		if blockStats.WrReqSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.block.stats",
				Desc:    "Number of write requests from a block device.",
				Metric:  "wr.req",
				Value:   uint64(blockStats.WrReq),
				Tags:    tags,
			})
		}

		if blockStats.WrTotalTimesSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.block.stats",
				Desc:    "Amount of time spent writing from a block device, in seconds.",
				Metric:  "wr.total.times",
				Value:   uint64(blockStats.WrTotalTimes),
				Tags:    tags,
			})
		}

		if blockStats.FlushReqSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.block.stats",
				Desc:    "Number of flush requests from a block device.",
				Metric:  "flush.req",
				Value:   uint64(blockStats.FlushReq),
				Tags:    tags,
			})
		}
		if blockStats.FlushTotalTimesSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.block.stats",
				Desc:    "Amount of time spent flushing of a block device, in seconds.",
				Metric:  "flush.total.times",
				Value:   uint64(blockStats.FlushTotalTimes),
				Tags:    tags,
			})
		}
		// Skip "Errs", as the documentation does not clearly
		// explain what this means.
	}

	// Report network interface statistics.
	for _, iface := range desc.Devices.Interfaces {
		interfaceStats, err := domain.InterfaceStats(iface.Target.Device)
		if err != nil {
			return points, err
		}
		/*
		 domifstat ubuntu1 vnet0
		vnet0 rx_bytes 1169845
		vnet0 rx_packets 10001
		vnet0 rx_errs 0
		vnet0 rx_drop 0
		vnet0 tx_bytes 0
		vnet0 tx_packets 0
		vnet0 tx_errs 0
		vnet0 tx_drop 0
		*/
		var tags map[string]string
		tags = map[string]string {}
		tags["Bridge"] = iface.Source.Bridge;
		tags["Device"] = iface.Target.Device;
		if interfaceStats.RxBytesSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.interface.stats",
				Desc:    "Number of bytes received on a network interface, in bytes.",
				Metric:  "rx.bytes",
				Value:   uint64(interfaceStats.RxBytes),
				Tags:    tags,
			})

		}
		if interfaceStats.RxPacketsSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.interface.stats",
				Desc:    "Number of packets received on a network interface.",
				Metric:  "rx.packets",
				Value:   uint64(interfaceStats.RxPackets),
				Tags:    tags,
			})

		}
		if interfaceStats.RxErrsSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.interface.stats",
				Desc:    "Number of packet receive errors on a network interface.",
				Metric:  "rx.errs",
				Value:   uint64(interfaceStats.RxErrs),
				Tags:    tags,
			})
		}
		if interfaceStats.RxDropSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.interface.stats",
				Desc:    "Number of packet receive drops on a network interface.",
				Metric:  "rx.drop",
				Value:   uint64(interfaceStats.RxDrop),
				Tags:    tags,
			})
		}

		if interfaceStats.TxBytesSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.interface.stats",
				Desc:    "Number of bytes transmitted on a network interface, in bytes.",
				Metric:  "tx.bytes",
				Value:   uint64(interfaceStats.TxBytes),
				Tags:    tags,
			})

		}
		if interfaceStats.TxPacketsSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.interface.stats",
				Desc:    "Number of packets transmitted on a network interface.",
				Metric:  "tx.packets",
				Value:   uint64(interfaceStats.TxPackets),
				Tags:    tags,
			})

		}
		if interfaceStats.TxErrsSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.interface.stats",
				Desc:    "Number of packet transmitted errors on a network interface.",
				Metric:  "tx.errs",
				Value:   uint64(interfaceStats.TxErrs),
				Tags:    tags,
			})
		}
		if interfaceStats.TxDropSet {
			points = append(points, backend.Point{
				KvmHost: kvm.Hostname,
				Domain:  domainName,
				Group:   "domain.interface.stats",
				Desc:    "Number of packet transmitted drops on a network interface.",
				Metric:  "tx.drop",
				Value:   uint64(interfaceStats.TxDrop),
				Tags:    tags,
			})
		}
	}

	return points, nil
}

func (kvm *KVM) Query(interval int, channel *chan []backend.Point) {
	conn, err := kvm.Connect() //libvirt.NewConnect(uri)
	if err != nil {
		return
	}
	defer conn.Close()

	// Use ListDomains() as opposed to using ListAllDomains(), as
	// the latter is unsupported when talking to a system using
	// libvirt 0.9.12 or older.
	domainIds, err := conn.ListDomains()
	if err != nil {
		return
	}
	for _, id := range domainIds {
		domain, err := conn.LookupDomainById(id)
		if err == nil {
			points,err := kvm.CollectDomain(domain)
			domain.Free()
			if err != nil {
				return
			}

			*channel <- points
		}
	}

	return
}

func (kvm *KVM) QueryHost(channel *chan backend.HostStuct) {

	conn, err := kvm.Connect() //libvirt.NewConnect(uri)
	if err != nil {
		return
	}
	defer conn.Close()

	var domains []backend.HostInfo
	// Use ListDomains() as opposed to using ListAllDomains(), as
	// the latter is unsupported when talking to a system using
	// libvirt 0.9.12 or older.
	domainIds, err := conn.ListDomains()
	if err != nil || len(domainIds) == 0 {
		return
	}
	for _, id := range domainIds {
		domain, err := conn.LookupDomainById(id)
		if err == nil {
			domainName, err := domain.GetName()
			if err != nil {
				domain.Free()
				return
			}
			id, err := domain.GetID()
			if err != nil {
				domain.Free()
				return
			}
			info, err := domain.GetInfo()

			if err != nil {
				domain.Free()
				return
			}
			domain.Free()
			domains = append(domains, backend.HostInfo{
				Id:   id,
				Name: domainName,
				Info: info,
			})

		}
	}

	var finderS backend.HostStuct
	finderS.Host = kvm.Hostname
	finderS.Infos = domains
	*channel <- finderS

}
