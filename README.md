# MLSQL-Deploy

## Build

```
make all
```

## Command

```shell
./mlsql-deploy --version
./mlsql-deploy --help
```

## Deploy MLSQL Engine

Use [juicefs](https://github.com/juicedata/juicefs) to create a file system, here we use Redis as meta storage.
You can download the juicefs tool from [JuiceFS 0.15.2](https://github.com/juicedata/juicefs/releases/tag/v0.15.2)

```shell
./juicefs format \
	--storage file \
	--bucket /tmp/jfs \
	redis://127.0.0.1:6379/1 \
	mlsql-k8s-storage
```


```shell
mlsql-deploy run \
--kube-config  /Users/allwefantasy/.kube/config \
--engine-name mlsql-k8s   \
--engine-image localhost:32000/mlsql-engine:3.0-2.1.0-SNAPSHOT   \
--engine-executor-core-num 2   \
--engine-executor-num 1   \
--engine-executor-memory 2048 \
--engine-driver-core-num 2   \
--engine-driver-memory 2048 \
--engine-access-token mlsql   \
--engine-jar-path-in-container local:///home/deploy/libs/streamingpro-mlsql-spark_3.0_2.12-2.1.0-SNAPSHOT.jar   \
--storage-name  mlsql-k8s-storage \
--storage-meta-url  redis://linux:6379/1
```