# ceph-metrics
A small service for Ceph to collect/send metrics (to graphite/influxdb) and http based alert checks

### Before you begin

Make sure you git submodules:

```
$ git submodule init
$ git submodule update
```

### Prepare your dev environment

```
$ vagrant up
$ cd provision
$ ansible-playbook -i inventory provision.yml
```

### Build

```
$ vagrant ssh ceph-metric-0
$ cd /vagrant/workspace/src/ceph-metrics
$ make install
```

### Run

```
$ ceph-metrics --help
```

Docker is also available:

```
$ docker build -t ceph-metrics .
```


#### TODO

- fix this README
- Clustering and HA needs work
- health checks and metrics needs improvement
- need docs
