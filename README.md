# Container Exporter
[Prometheus](https://github.com/prometheus/prometheus) exporter exposing container metrics.

The container-exporter requests a list of containers running on the host by talking to a 
container manager. Right now, Docker as container manager is supported.
It then gathers various container metrics by using [libcontainer](https://github.com/docker/libcontainer)
and [DockerClient](https://github.com/fsouza/go-dockerclient) and then exposes them for prometheus' consumption. 

## Run it as container

    docker run -p 9104:9104 -v /sys/fs/cgroup:/cgroup \
               -v /var/run/docker.sock:/var/run/docker.sock prom/container-exporter

## Support for labels

Specify all Docker label whose values you would like to tag your Prometheus metrics with by using the -labels parameter to the container exporter binary (or docker container). For example if you have a container labeled with LabelA and LabelB and a second container labeled with LabelB and LabelC as shown below. You can launch container exporter with the parameter -labels=LabelA,LabelB,LabelC. 

    docker run --name ContainerA --label LabelA=ValueA --label LabelB=ValueB [IMAGE] 
    docker run --name ContainerB --label LabelB=ValueB --label LabelC=ValueC [IMAGE] 
    docker run -p 9104:9104 -v /sys/fs/cgroup:/cgroup \
               -v /var/run/docker.sock:/var/run/docker.sock prom/container-exporter -labels=LabelA,LabelB,LabelC
               
This will load to the metrics shown below. Note that an empty string is reported for any container that does not define a label that is specified to container exporter.

    container_cpu_throttled_periods_total{LabelA="ValueA",LabelB="ValueB",LabelC="",name="ContainerA"...
    container_cpu_throttled_periods_total{LabelA="",LabelB="ValueB",LabelC="ValueC",name="ContainerB"...
