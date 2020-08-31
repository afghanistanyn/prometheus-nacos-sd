# prometheus-nacos-sd

Prometheus service discovery using nacos and `file_sd_config`.

----
## Install

### Precompiled binaries

Download from https://github.com/afghanistanyn/prometheus-nacos-sd/releases


### manual build
```
git clone https://github.com/afghanistanyn/prometheus-nacos-sd
cd prometheus-nacos-sd
make buildx
```

### Docker

```
docker pull afghanistanyn/prometheus-nacos-sd:v1.0
docker run -it -d -v /tmp:/tmp afghanistanyn/prometheus-nacos-sd --nacos.address=192.168.1.1:8848 --nacos.namespace=dev --output.file=/tmp/nacos_sd_dev.json
ls /tmp/nacos_sd_dev.json
```


----
## Usage

- run 

    ```
    ./prometheus-nacos-sd --nacos.address=192.168.1.1:8848 --nacos.namespace=dev nacos.group=DEFAULT_GROUP --output.file=nacos_sd_dev.json --refresh.interval=30
    ```

- default parameters
```
   nasos.address=localhost:8848
   nacos.namespace=public
   nacos.group=DEFAULT_GROUP
   refresh.interval=60

```

---- 
## Generated json format

generated json should be follow prometheus `file_sd_config` format like below:

```json
[
    {
        "targets": [
            "192.168.1.1:5066"
        ],
        "labels": {
            "__meta_context_path": "/miniapp-gateway",
            "__meta_nacos_group": "DEFAULT_GROUP",
            "__meta_nacos_namespace": "dev",
            "__meta_preserved_register_source": "SPRING_CLOUD",
            "__metrics_path__": "/miniapp-gateway/actuator/prometheus",
            "job": "miniapp-gateway"
        }
    },
    {
        "targets": [
            "192.168.1.1:5064"
        ],
        "labels": {
            "__meta_context_path": "/",
            "__meta_nacos_group": "DEFAULT_GROUP",
            "__meta_nacos_namespace": "dev",
            "__meta_preserved_register_source": "SPRING_CLOUD",
            "__metrics_path__": "/actuator/prometheus",
            "job": "web-gateway"
        }
    }
]
```

## about prometheus metric path
```
we set __metrics_path__ to "/actuator/prometheus" by default.
but if you set metadata with key 'context_path' in your application metadata like below , we will rewrite it to "content_path/actuator/prometheus"
```

```yaml
spring:
  cloud:
    nacos:
      discovery:
        server-addr: 192.168.1.1:8848
        namespace: dev
        metadata:
          context_path: ${server.servlet.context-path:/}
```



## Example prometheus settings

The part of your `prometheus.yml` is probably as follows.

```yaml
  scrape_configs:
      - job_name: 'nacos-discorvery'
        file_sd_configs:
        - files:
          - /apps/prometheus/conf/nacos_sd_dev.json
          - /apps/prometheus/conf/nacos_sd_test.json
          refresh_interval: 1m
    
        relabel_configs:
        - regex: 'preserved_register_source'
          action: labeldrop
```

