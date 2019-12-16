# AWX Exporter

AWX Exporter is used to create Prometheus/AlertManager/Blackbox
configuration, from the Ansible AWX groups and hosts. The Prometheus
configurations can be set for the whole group of host or with override
option for a single host. The result can be added to prometheus as
a file target source. The blackbox configuration are set for the given
host and AlertManager is set for the given group. It also can be 
overridden by the given host settings.  

## Config

AWX Exporter always looks for the configuration in the same directory
or the path of the --config. It should be an `ini` file and following
parameters should be set.


```lang=ini
[AWX]
HostName='https://host'
UserName='admin'
Token=''
InventorySources=''
TimeOut = 10s

[PROMETHEUS]
ConfigName='prometheus_config' #Should be set in group or host in AWX
ConfigHostOverride=True
HostNameVar='cmdb_name' #Should be set in host in AWX
IpVar='ansible_host' #Should be set in host in AWX

[ALERTMANAGER]
ConfigName='alertmanager_config' #Should be set in group in AWX
SourceFile='/etc/alertmanager/alertmanager.yml'
RequireTLSDefault=False
SendResolveDefault=True

[BLACKBOX]
ConfigName='blackbox_config' #Should be set in host in AWX
IgnoredGroups='cmdb_imported,guests'
HostNameVar='cmdb_name'   (Should be set in host in AWX)
IpVar='ansible_ssh_host'  (Should be set in host in AWX)
```

In Awx you need to also have the given variables used so the data can
be generated without problem.

For each of the exporter the following syntax should be used:

- Prometheus (In host or group): 

```lang=yaml
prometheus_config:
  - name: wmi
    exporter: vmi-exporter
    port: 9182
```

- Blackbox (In host):

```lang=yaml
blackbox_config:
  - module: http_2xx
    targets:
      - 'https://exampel1.com/'
      - 'https://example2.com'
```

- AlertManager (In group):

```lang=yaml
alertmanager_config:
  - name: admin-receiver
    type: email
    receiver-config:
      to: admin@admin.com
```
## Running

To run the application simply copy the binary in the right directory,
set the config.ini and run it like any other application.

```lang=bash
# Prometheus Mode
./awx-exporter -prometheus -config-path="config.ini"
# AlertManager Mode
./awx-exporter -alertmanager -config-path="config.ini"
# Blackbox Mode
./awx-exporter -blackbox -config-path="config.ini"
```

The result will be written on stdout. Upon errors the program
will break with Fatal status.

## License

See LICENSE file.

## Change Log

See CHANGELOG file.