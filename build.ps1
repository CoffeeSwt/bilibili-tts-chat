# Simple Go Project Build Script
# Build bilibili-tts-chat project to dist directory

Write-Host "=== Go Project Build Script ===" -ForegroundColor Green
Write-Host ""

# Check Go environment
Write-Host "Checking Go environment..." -ForegroundColor Cyan
try {
    $goVersion = go version 2>$null
    if (-not $goVersion) {
        Write-Host "Error: Go environment not found, please install Go first" -ForegroundColor Red
        exit 1
    }
    Write-Host "✓ Go environment check passed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "Error: Go environment check failed" -ForegroundColor Red
    exit 1
}

# Clean and create dist directory
Write-Host "Preparing dist directory..." -ForegroundColor Cyan
if (Test-Path "dist") {
    Remove-Item "dist" -Recurse -Force
    Write-Host "✓ Cleaned old dist directory" -ForegroundColor Green
}

New-Item -ItemType Directory -Path "dist" -Force | Out-Null
Write-Host "✓ Created dist directory" -ForegroundColor Green

# Build binary
Write-Host "Building binary..." -ForegroundColor Cyan
try {
    $buildResult = go build -o "dist\bilibili-tts-chat.exe" . 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: Build failed: $buildResult" -ForegroundColor Red
        exit 1
    }
    
    if (Test-Path "dist\bilibili-tts-chat.exe") {
        $fileSize = (Get-Item "dist\bilibili-tts-chat.exe").Length
        $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
        Write-Host "✓ Build successful: bilibili-tts-chat.exe (${fileSizeMB}MB)" -ForegroundColor Green
    } else {
        Write-Host "Error: Build completed but executable not found" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "Error: Build process failed: $_" -ForegroundColor Red
    exit 1
}

# Copy configuration files
Write-Host "Copying and preparing configuration files..." -ForegroundColor Cyan

$configFiles = @(
    @{Source = ".env.example"; Target = ".env"}
)

foreach ($config in $configFiles) {
    $sourceFile = $config.Source
    $targetFile = $config.Target
    
    if (Test-Path $sourceFile) {
        try {
            # Copy the .example file to dist directory
            Copy-Item $sourceFile -Destination "dist\"
            Write-Host "✓ Copied: $sourceFile to dist/" -ForegroundColor Green
            
            # Rename the file in dist directory to remove .example suffix
            $distExampleFile = "dist\$sourceFile"
            $distTargetFile = "dist\$targetFile"
            
            if (Test-Path $distExampleFile) {
                Rename-Item $distExampleFile -NewName $targetFile
                Write-Host "✓ Renamed: $sourceFile → $targetFile" -ForegroundColor Green
            } else {
                Write-Host "Warning: Could not find copied file: $distExampleFile" -ForegroundColor Yellow
            }
        } catch {
            Write-Host "Error: Failed to process $sourceFile : $_" -ForegroundColor Red
        }
    } else {
        Write-Host "Warning: Source file not found: $sourceFile" -ForegroundColor Yellow
    }
}

# Show results
Write-Host ""
Write-Host "=== Build Completed ===" -ForegroundColor Green
Write-Host "Build directory: $(Get-Location)\dist" -ForegroundColor White
Write-Host ""
Write-Host "Files in dist directory:" -ForegroundColor Yellow
Get-ChildItem "dist" | ForEach-Object {
    $size = if ($_.PSIsContainer) { "DIR" } else { "$([math]::Round($_.Length / 1KB, 1))KB" }
    Write-Host "  $($_.Name) ($size)" -ForegroundColor White
}

Write-Host ""
Write-Host "Usage:" -ForegroundColor Yellow
Write-Host "  1. Edit .env file and configure your environment variables" -ForegroundColor White
Write-Host "  2. Edit config.yaml file and configure your application settings" -ForegroundColor White
Write-Host "  3. Run: .\bilibili-tts-chat.exe" -ForegroundColor White