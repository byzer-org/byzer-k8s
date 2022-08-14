# Byzer-K8S

byzer-k8s provides a CLI to deploy Byzer Engine on K8S. There are three steps:

- Building byzerup from Source
- Setup JuiceFS File System 
- Running byzerup CLI

## Build

```
make all
```

## Command

```shell
mv ./byzerup-linux-amd64 ./byzerup
./byzerup --version
./byzerup --help
```

It's recommended to install JuiceFS to provide a distributes file system. Together with K8S, it provides user a local
cloud mimic environment.

## [JuiceFS](https://github.com/juicedata/juicefs) File System Setup

JuiceFS needs an object storage and database to startup.

### Object Storage Quick Setup (Optional)

JuiceFS supports [a variety of object storages](https://github.com/juicedata/juicefs#supported-object-storage), here.
Here's an example start a local [MinIO](https://github.com/minio/minio) instance

```shell
mkdir -p ~/minio-data
docker run -d --name minio \
        -v ~/minio-data:/data \
        -p 9000:9000 \
        -p 9001:9001 \
        --restart unless-stopped \
        minio/minio server /data --console-address ":9001"
 
```

Login [MinIO cosnole](http://127.0.0.1:9001/) and create a bucket named `byjfs`

### Meta Storage Setup

A quick way to start Redis instance:

```shell
docker run -d --name redis -p 6379:6379 redis
```

Please note that MySQL/SQLite is supported as well.

### JuiceFS Initialization

Please download and untar [JuiceFS 0.15.2](https://github.com/juicedata/juicefs/releases/tag/v0.15.2), then run the
following
command:

```shell
## Please go to untarred directory and run this command 
sudo install juicefs /usr/local/bin
## Check JuiceFS installation
juicefs --version
## Format JuiceFS
juicefs format \
  --storage minio \
  --bucket http://127.0.0.1:9000/jfs \
  --access-key minioadmin \
  --secret-key minioadmin \
  redis://127.0.0.1:6379/1 \
  byjfs
```

## Deploying Byzer Engine on k8s

```shell
./byzerup run \
--kube-config  /Users/allwefantasy/.kube/config \
--engine-name byzer-engine-william   \
--engine-version 2.3.1 \
--engine-service-account-name byzer \
--engine-namespace byzer-ns \
--engine-executor-core-num 2   \
--engine-executor-num 2   \
--engine-executor-memory 2048 \
--engine-driver-core-num 2   \
--engine-driver-memory 4048 \
--engine-access-token mlsql  \
--storage-name  byjfs \
--storage-meta-url  "redis://127.0.0.1:6379/1" \
--engine-config ./.engine.config
```

.engine.config example:

```Shell 
engine.spark.kubernetes.container.image.pullPolicy=IfNotPresent 
engine.streaming.datalake.path=./data/
```

Done.
