# Go项目构建脚本
# 构建bilibili-tts-chat项目到dist目录

Write-Host "=== Go项目构建脚本 ===" -ForegroundColor Green
Write-Host ""

# 检查Go环境
Write-Host "正在检查Go环境..." -ForegroundColor Cyan
try {
    $goVersion = go version 2>$null
    if (-not $goVersion) {
        Write-Host "错误：未找到Go环境，请先安装Go" -ForegroundColor Red
        exit 1
    }
    Write-Host "Go环境检查通过：$goVersion" -ForegroundColor Green
} catch {
    Write-Host "错误：Go环境检查失败" -ForegroundColor Red
    exit 1
}

# 清理并创建dist目录
Write-Host "正在准备dist目录..." -ForegroundColor Cyan
if (Test-Path "dist") {
    Remove-Item "dist" -Recurse -Force
    Write-Host "已清理旧的dist目录" -ForegroundColor Green
}

New-Item -ItemType Directory -Path "dist" -Force | Out-Null
Write-Host "已创建dist目录" -ForegroundColor Green

# 构建二进制文件
Write-Host "正在构建二进制文件..." -ForegroundColor Cyan
try {
    $buildResult = go build -o "dist\bilibili-tts-chat.exe" . 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "错误：构建失败：$buildResult" -ForegroundColor Red
        exit 1
    }
    
    if (Test-Path "dist\bilibili-tts-chat.exe") {
        $fileSize = (Get-Item "dist\bilibili-tts-chat.exe").Length
        $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
        Write-Host "构建成功：bilibili-tts-chat.exe ($fileSizeMB MB)" -ForegroundColor Green
    } else {
        Write-Host "错误：构建完成但未找到可执行文件" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "错误：构建过程失败：$_" -ForegroundColor Red
    exit 1
}

# 复制配置文件
Write-Host "正在复制和准备配置文件..." -ForegroundColor Cyan

$configFiles = @(
    @{Source = ".env.example"; Target = ".env"}
    @{Source = "user.example.json"; Target = "user.json"}
    @{Source = "voices.json"; Target = "voices.json"}
)

foreach ($config in $configFiles) {
    $sourceFile = $config.Source
    $targetFile = $config.Target
    
    if (Test-Path $sourceFile) {
        try {
            if ($sourceFile -eq ".env.example") {
                # 复制.example文件到dist目录
                Copy-Item $sourceFile -Destination "dist\"
                Write-Host "已复制：$sourceFile 到 dist/" -ForegroundColor Green
                
                # 在dist目录中重命名文件以移除.example后缀
                $distExampleFile = "dist\$sourceFile"
                $distTargetFile = "dist\$targetFile"
                
                if (Test-Path $distExampleFile) {
                    Rename-Item $distExampleFile -NewName $targetFile
                    Write-Host "已重命名：$sourceFile -> $targetFile" -ForegroundColor Green
                } else {
                    Write-Host "警告：找不到已复制的文件：$distExampleFile" -ForegroundColor Yellow
                }
            } elseif ($sourceFile -eq "user.example.json") {
                # 复制.example文件到dist目录
                Copy-Item $sourceFile -Destination "dist\"
                Write-Host "已复制：$sourceFile 到 dist/" -ForegroundColor Green
                
                # 在dist目录中重命名文件以移除.example后缀
                $distExampleFile = "dist\$sourceFile"
                $distTargetFile = "dist\$targetFile"
                
                if (Test-Path $distExampleFile) {
                    Rename-Item $distExampleFile -NewName $targetFile
                    Write-Host "已重命名：$sourceFile -> $targetFile" -ForegroundColor Green
                } else {
                    Write-Host "警告：找不到已复制的文件：$distExampleFile" -ForegroundColor Yellow
                }
            } else {
                # 直接复制其他配置文件
                Copy-Item $sourceFile -Destination "dist\"
                Write-Host "已复制：$sourceFile 到 dist/" -ForegroundColor Green
            }
        } catch {
            Write-Host "错误：处理 $sourceFile 失败：$_" -ForegroundColor Red
        }
    } else {
        Write-Host "警告：找不到源文件：$sourceFile" -ForegroundColor Yellow
    }
}

# 显示结果
Write-Host ""
Write-Host "=== 构建完成 ===" -ForegroundColor Green
Write-Host "构建目录：$(Get-Location)\dist" -ForegroundColor White
Write-Host ""
Write-Host "dist目录中的文件：" -ForegroundColor Yellow
Get-ChildItem "dist" | ForEach-Object {
    $size = if ($_.PSIsContainer) { "目录" } else { "$([math]::Round($_.Length / 1KB, 1))KB" }
    Write-Host "  $($_.Name) ($size)" -ForegroundColor White
}

Write-Host ""
Write-Host "使用方法：" -ForegroundColor Yellow
Write-Host "  1. 编辑.env文件并配置您的环境变量" -ForegroundColor White
Write-Host "  2. 编辑user.json和voices.json文件并配置您的应用设置" -ForegroundColor White
Write-Host "  3. 运行：.\bilibili-tts-chat.exe" -ForegroundColor White