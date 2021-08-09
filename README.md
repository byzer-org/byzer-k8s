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

## Deploy MLSQL Engine on K8s

You need first downloading [JuiceFS 0.15.2](https://github.com/juicedata/juicefs/releases/tag/v0.15.2) then run following
command:

```shell
./juicefs format \
	--storage file \
	--bucket /tmp/jfs \
	redis://192.168.31.95:6379/1 \
	mlsql-k8s-storage
```

Notice that you should replace the `file` with really object storage or HDFS.

Now you can deploy MLSQL Engine on k8s:

```shell
./mlsql-deploy run \
--kube-config  /Users/allwefantasy/.kube/config \
--engine-name mlsql-k8s   \
--engine-image localhost:32000/mlsql-engine:3.0-2.1.0-SNAPSHOT   \
--engine-executor-core-num 2   \
--engine-executor-num 1   \
--engine-executor-memory 2048 \
--engine-driver-core-num 2   \
--engine-driver-memory 2048 \
--engine-access-token mlsql   \
--engine-jar-path-in-container local:///home/deploy/mlsql/mlsql-engine_3.0-2.1.0-SNAPSHOT/libs/streamingpro-mlsql-spark_3.0_2.12-2.1.0-SNAPSHOT.jar   \
--storage-name  mlsql-k8s-storage \
--storage-meta-url  redis://192.168.31.95:6379/1
```

Done.