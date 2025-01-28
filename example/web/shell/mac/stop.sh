#!/bin/bash

# 定义要关闭的进程名数组
processes=("nats-server" "service" "attach" "alarm" "notice")

# 关闭所有指定进程
for process in "${processes[@]}"; do
    echo "尝试关闭 $process 进程..."
    if pkill -TERM "$process"; then
        echo "$process 进程已发送终止信号。"
    else
        echo "$process 进程未找到，无需关闭。"
    fi
done
