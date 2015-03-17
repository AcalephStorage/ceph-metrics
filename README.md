# ceph-metrics
A small service for Ceph to collect/send metrics (to graphite/influxdb) and http based alert checks

To prepare dev environment:

```
$ vagrant up
$ cd provision
$ ansible-playbook -i inventory provision.yml
```

To build:

```
$ vagrant ssh ceph-metric-0
$ cd /vagrant/workspace/src/ceph-metrics
$ make install
```

To run:

```
$ ceph-metrics --help
```

Docker is also available:

```
$ docker build -t ceph-metrics .
```


TODO:
- fix this README
- Clustering and HA needs work
- health checks and metrics needs improvement
- need docs
