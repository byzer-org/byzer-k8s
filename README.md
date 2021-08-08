# MLSQL-Deploy

## Build

```
make all
```

## Run

```shell
./mlsql-deploy --version
./mlsql-deploy --help
```

## Example

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
--storage-meta-url  redis://linux:6379/0
```