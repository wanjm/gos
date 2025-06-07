#!/bin/bash

# 定义文件名数组
files=("$@")

# 如果数组为空，设置默认值
if [ ${#files[@]} -eq 0 ]; then
    files=("project" "package" "struct"
        "interface" "function" "method" "variable" "gosourse"
    )
fi

# 遍历数组检查并创建文件
for file in "${files[@]}"; do
    file="$file.go"
    if [ ! -f "astinfo/$file" ]; then
        echo "Creating astinfo/$file"
        echo "package astinfo" > "astinfo/$file"
    fi
done
