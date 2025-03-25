#!/bin/bash

# 定义支持的架构
ARCH_CHOICES=("amd64" "arm64")
ARCH_MAP=("amd64"="x86_64" "arm64"="ARM64")

# 显示架构选择菜单
echo "请选择目标架构："
for i in "${!ARCH_CHOICES[@]}"; do
  echo "$((i+1)): ${ARCH_CHOICES[$i]}"
done

# 读取用户输入
read -p "输入编号 (1 或 2): " ARCH_CHOICE

# 验证用户输入
if [[ $ARCH_CHOICE -lt 1 || $ARCH_CHOICE -gt ${#ARCH_CHOICES[@]} ]]; then
  echo "无效的选择！退出脚本。"
  exit 1
fi

# 获取用户选择的架构
SELECTED_ARCH=${ARCH_CHOICES[$((ARCH_CHOICE-1))]}
ARCH_NAME=${ARCH_MAP[$SELECTED_ARCH]}

# 设置目标操作系统
OS="linux"

# 显示选择结果
echo "你选择了目标架构: $SELECTED_ARCH ($ARCH_NAME)"

# 编译
echo "正在编译为 $OS/$SELECTED_ARCH ..."
GOOS=$OS GOARCH=$SELECTED_ARCH go build -o socks5-proxy-$OS-$SELECTED_ARCH main.go

# 检查编译是否成功
if [ $? -ne 0 ]; then
  echo "编译失败！请检查错误信息。"
  exit 1
fi

# 打包
echo "正在打包为 $OS-$ARCH_NAME 格式..."
TAR_FILE="socks5-proxy-$OS-$SELECTED_ARCH.tar.gz"
tar -czvf $TAR_FILE socks5-proxy-$OS-$SELECTED_ARCH

# 检查打包是否成功
if [ $? -ne 0 ]; then
  echo "打包失败！请检查错误信息。"
  exit 1
fi

# 清理
echo "清理临时文件..."
rm socks5-proxy-$OS-$SELECTED_ARCH

# 显示完成信息
echo "编译和打包完成！"
echo "输出文件: $TAR_FILE"