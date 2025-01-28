#!/bin/bash

# 关闭所有指定进程
pkill -TERM nats-server
pkill -TERM service
pkill -TERM attach
pkill -TERM alarm
pkill -TERM notice

# 检查是否还有残留进程
echo "剩余进程检查："
pgrep -l 'nats-server|service|attach|alarm|notice' || echo "所有进程已关闭"