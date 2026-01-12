# FilePath    : go-utils\run.ps1
# Author      : jiaopengzi
# Blog        : https://jiaopengzi.com
# Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
# Description : 运行脚本，提供代码格式化、单元测试、go lint 和 gopls check 功能 

# 显示菜单
Write-Host ""
Write-Host "请选择需要执行的命令："
Write-Host "  1 - 格式化代码"
Write-Host "  2 - 单元测试"
Write-Host "  3 - go lint"
Write-Host "  4 - gopls check"
Write-Host ""

# 接收用户输入的操作编号
$choice = Read-Host "请输入编号选择对应的操作"
Write-Host ""

# 格式化代码,
function formatCode {
    go fmt ./...
}

# 单元测试
function test {
    go test -v ./...
}

# 使用 golangci-lint run 命令检查代码格式和静态错误
function goLint {
    go vet ./...
    golangci-lint run
    Write-Host "✅ 代码格式和静态检查完毕"
}

# gopls check 检查代码格式和静态错误
function goplsCheck {
    # 运行前添加策略 Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
    # 这个脚本使用 gopls check 检查当前目录及其子目录中的所有 Go 文件。
    # 主要是在 gopls 升级后或者go版本升级后检查代码是否有问题.

    # 拿到当前目录下所有的 .go 文件数量
    $goFilesCount = Get-ChildItem -Path . -Filter *.go -File -Recurse | Measure-Object | Select-Object -ExpandProperty Count

    # 每分钟大约处理文件为 26 个, 计算出大概所需时间(秒)
    $estimatedTime = [Math]::Round($goFilesCount / 26 * 60)

    # 获取当前目录及其子目录中的所有 .go 文件
    $goFiles = Get-ChildItem -Recurse -Filter *.go

    # 记录开始时间
    $startTime = Get-Date

    # 设置定时器间隔
    $interval = 60

    # 初始化已检查文件数量
    $checkedFileCount = 0

    # 初始化上次输出时间
    $lastOutputTime = $startTime

    # 遍历每个 .go 文件并运行 gopls check 命令
    Write-Host "正在检查, 耗时预估 $estimatedTime 秒, 请耐心等待..." -ForegroundColor Green
    foreach ($file in $goFiles) {
        # Write-Host "正在检查 $($file.FullName)..."
        gopls check $file.FullName
        if ($LASTEXITCODE -ne 0) {
            Write-Host "检查 $($file.FullName) 时出错" -ForegroundColor Red
        } 
        $checkedFileCount++

        # 获取当前时间
        $currentTime = Get-Date
        $elapsedTime = $currentTime - $startTime

        # 检查是否已经超过了设定的时间间隔
        if (($currentTime - $lastOutputTime).TotalSeconds -ge $interval) {
            $roundedElapsedTime = [Math]::Round($elapsedTime.TotalSeconds)
            Write-Host "当前已耗时 $roundedElapsedTime 秒, 已检查文件数量: $checkedFileCount" -ForegroundColor Yellow
            # 更新上次输出时间
            $lastOutputTime = $currentTime
        }
    }

    # 记录结束时间
    $endTime = Get-Date

    # 计算耗时时间
    $elapsedTime = $endTime - $startTime

    # 显示总耗时时间和总文件数量
    $roundedElapsedTime = [Math]::Round($elapsedTime.TotalSeconds)
    Write-Host "检查结束, 总耗时 $roundedElapsedTime 秒, 总文件数量: $($goFiles.Count), 已检查文件数量: $checkedFileCount" -ForegroundColor Green
}

# switch 要放到最后 
# 执行用户选择的操作
switch ($choice) {
    1 { formatCode }
    2 { test }
    3 { goLint }
    4 { goplsCheck }
    default { Write-Host "❌ 无效的选择" }
}