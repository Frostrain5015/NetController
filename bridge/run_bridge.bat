@echo off
REM NetController SSH Bridge 启动脚本
REM 密码通过环境变量传入，不写入命令行历史

set NC_SSH_HOST=116.62.179.231
set NC_SSH_USER=root
set /p NC_SSH_PASSWORD="Enter SSH password: "

python bridge.py --host %NC_SSH_HOST% --user %NC_SSH_USER% --password %NC_SSH_PASSWORD% --config config.yaml
