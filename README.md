# 抓取kvm指标信息

## 依赖
[libvirt](https://libvirt.org/)
[libvirt-go](https://github.com/libvirt/libvirt-go)

## 要求

远程抓取时候kvm打开tcp侦听

## 抓取指标如下

```
libvirt_domain_block_stats_read_bytes_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_read_requests_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_read_total_time{domain="...",source_file="...",target_device="..."}

libvirt_domain_block_stats_flush_total_time{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_flush_requests_total{domain="...",source_file="...",target_device="..."}

libvirt_domain_block_stats_write_bytes_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_write_requests_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_write_total_time{domain="...",source_file="...",target_device="..."}

libvirt_domain_info_cpu_time_seconds_total{domain="..."}
libvirt_domain_info_maximum_memory_bytes{domain="..."}
libvirt_domain_info_memory_usage_bytes{domain="..."}
libvirt_domain_info_virtual_cpus{domain="..."}

libvirt_domain_interface_stats_receive_bytes_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_receive_drops_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_receive_errors_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_receive_packets_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_bytes_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_drops_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_errors_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_packets_total{domain="...",source_bridge="...",target_device="..."}

```


## docker

Dockerfile2 是原生模式打出来的镜像有1.5G    
Dockerfile 是我处理以后的，打出来镜像80M

请开启kvm的tcp连接，端口16509

```shell
  
 docker build -t leapar/libvirt:0.11 .
 docker run -d leapar/libvirt:0.11
 docker run -v `pwd`/libvirt-metrics.json:/etc/libvirt-metrics.json -d test/libvirt:0.1  

```
