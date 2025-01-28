#!/bin/bash

set -e  # 任何命令失败时立即退出脚本

cd ./nats/*/ && nohup ./nats-server >./nats.log &
cd ./service && nohup ./service >./service.log &
cd ./attach && nohup ./attach > ./attach.log &
cd ./alarm && nohup ./alarm >./alarm.log &
cd ./notice && nohup ./notice -address=0.0.0.0:18003 -nats=127.0.0.1:4222 >./notice.log &

# 统一检查进程状态
pgrep -l 'nats-server|service|attach|alarm|notice'