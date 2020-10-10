#!/bin/bash

export HBASE_HEAPSIZE="4G"
export HBASE_OPTS="-XX:+UseParallelGC -XX:+UseParallelOldGC"
export SERVER_GC_OPTS="-XX:+PrintGCDetails -XX:+PrintGCDateStamps -Xloggc:/work/logs/gc.log -XX:+UseGCLogFileRotation -XX:GCLogFileSize=128m"
export HBASE_ROOT_LOGGER="INFO,DRFA"
export HBASE_SECURITY_LOGGER="INFO,RFAS"
export HBASE_IDENT_STRING=$CONTAINER_ID

mkdir -p /work/conf
ln -s /configmaps/hbase-site.xml /work/conf/
ln -s /configmaps/log4j.properties /work/conf/
ln -s /configmaps/hadoop-metrics2-hbase.properties /work/conf/

useradd hbase -u 10086

mkdir -p /work/all-logs/$CONTAINER_ID
ln -s /work/all-logs/$CONTAINER_ID /work/logs
chown -R hbase:hbase /work/all-logs/$CONTAINER_ID

#./bin/hbase regionserver
su hbase -c "/work/bin/hbase master start"

#sleep 3600

