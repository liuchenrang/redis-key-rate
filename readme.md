# 无埋点，统计key的命中率

1. tcpdupm -i lo0 tcp and port 6379 -a /tmp/redis.pcap
2. ./main -r redis.pacp -keyExp $key

