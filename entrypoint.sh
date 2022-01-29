#!/bin/sh

redisAddress=${redisAdress:-""}

if [ "${redisAddress}" = "" ]; then
    redisAddress="localhost:6379"
    mkdir /var/lib/redis
    redis-server /etc/redis/redis.conf
fi

/ospserver \
    --port=${port:-443} \
    --redis-address=${redisAddress} \
    --redis-db=${redisDB:-0} \
    --redis-password=${redisPassword:-""} \
    --cert-file=${certFile:-""}\
    --cert-key-file=${certKeyFile:-""}
