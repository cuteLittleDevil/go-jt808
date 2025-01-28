#!/bin/bash

set -e  # 任何命令失败时立即退出脚本

chmod -R +x ./

# 启动 nats-server
cd ./nats/*/ && nohup ./nats-server >./nats.log 2>&1 &

sleep 1
# 启动 service
cd ./service && nohup ./service >./service.log 2>&1 &

# 启动 attach
cd ./attach && nohup ./attach > ./attach.log 2>&1 &

# 启动 alarm
cd ./alarm && nohup ./alarm >./alarm.log 2>&1 &

# 启动 notice
cd ./notice && nohup ./notice -address=0.0.0.0:18003 -nats=127.0.0.1:4222 >./notice.log 2>&1 &

# 统一检查进程状态
# 使用 pgrep 检查进程是否存在
if pgrep -x -q 'nats-server'; then
    echo "nats-server 进程正在运行，PID: $(pgrep 'nats-server')"
else
    echo "nats-server 进程未运行"
fi

if pgrep -x -q 'service'; then
    echo "service 进程正在运行，PID: $(pgrep 'service')"
else
    echo "service 进程未运行"
fi

if pgrep -x -q 'attach'; then
    echo "attach 进程正在运行，PID: $(pgrep 'attach')"
else
    echo "attach 进程未运行"
fi

if pgrep -x -q 'alarm'; then
    echo "alarm 进程正在运行，PID: $(pgrep 'alarm')"
else
    echo "alarm 进程未运行"
fi

if pgrep -x -q 'notice'; then
    echo "notice 进程正在运行，PID: $(pgrep 'notice')"
else
    echo "notice 进程未运行"
fi
