# MLSQL-Deploy
mlsql-deploy provides a CLI to deploy MLSQL Engine on K8S. There are three steps:
- Building mlsql-deploy from Source
- JuiceFS File System Setup
- Running mlsql-deploy CLI

## Build
```
make all
```

## Command
```shell
./mlsql-deploy --version
./mlsql-deploy --help
```

It's recommended to install JuiceFS to provide a distributes file system. Together with K8S, it provides user a local
cloud mimic environment.

## [JuiceFS](https://github.com/juicedata/juicefs) File System Setup
JuiceFS needs an object storage and database to startup.

###  Object Storage Quick Setup
JuiceFS supports [a variety of object storages](https://github.com/juicedata/juicefs#supported-object-storage), here. Here's an example start a local [MinIO](https://github.com/minio/minio) instance
```shell
mkdir -p ~/minio-data
docker run -d --name minio \
        -v ~/minio-data:/data \
        -p 9000:9000 \
        -p 9001:9001 \
        --restart unless-stopped \
        minio/minio server /data --console-address ":9001"
 
```
Login [MinIO cosnole](http://127.0.0.1:9001/) and create a bucket named jfs
### Meta Storage Setup
A quick way to start Redis instance:
```shell
docker run -d --name redis -p 6379:6379 redis
```
Please note that MySQL/SQLite is supported as well.

### JuiceFS Initialization
Please download and untar [JuiceFS 0.15.2](https://github.com/juicedata/juicefs/releases/tag/v0.15.2), then run the following
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
  jfs
```

## Deploying MLSQL Engine on k8s
```shell
## K8S config file resides in ~/.kube/config by default.
## chncaesar/mlsql-engine-k8s:3.0-2.1.0-SNAPSHOT is a pre-built K8S image
./mlsql-deploy run \
  --kube-config  ~/.kube/config \
  --engine-name mlsql-k8s   \
  --engine-image chncaesar/mlsql-engine-k8s:3.0-2.1.0-SNAPSHOT \
  --engine-executor-core-num 2   \
  --engine-executor-num 1   \
  --engine-executor-memory 2048 \
  --engine-driver-core-num 2   \
  --engine-driver-memory 2048 \
  --engine-access-token mlsql   \
  --engine-jar-path-in-container local:///home/deploy/mlsql/mlsql-engine_3.0-2.1.0-SNAPSHOT/libs/streamingpro-mlsql-spark_3.0_2.12-2.1.0-SNAPSHOT.jar   \
  --storage-name  jfs \
  --storage-meta-url redis://127.0.0.1:6379/1
```

Done.