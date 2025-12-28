# Go Project Build Script
# Build bilibili-tts-chat to dist directory

Write-Host "=== Go Project Build Script ===" -ForegroundColor Green
Write-Host ""

# XOR Encryption Function
function Encrypt-String {
    param (
        [string]$InputString
    )
    
    if ([string]::IsNullOrEmpty($InputString)) {
        return ""
    }
    
    $Key = "bilibili-tts-chat-secret-key-2025"
    $InputBytes = [System.Text.Encoding]::UTF8.GetBytes($InputString)
    $EncryptedBytes = New-Object byte[] $InputBytes.Length
    
    # Convert Key to byte array to ensure correct XOR operation
    $KeyBytes = [System.Text.Encoding]::UTF8.GetBytes($Key)
    
    for ($i = 0; $i -lt $InputBytes.Length; $i++) {
        $EncryptedBytes[$i] = $InputBytes[$i] -bxor $KeyBytes[$i % $KeyBytes.Length]
    }
    
    return [Convert]::ToBase64String($EncryptedBytes)
}

# Check Go Environment
Write-Host "Checking Go environment..." -ForegroundColor Cyan
try {
    $goVersion = go version 2>$null
    if (-not $goVersion) {
        Write-Host "Error: Go not found, please install Go first" -ForegroundColor Red
        exit 1
    }
    Write-Host "Go environment check passed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "Error: Go environment check failed" -ForegroundColor Red
    exit 1
}

# Check Wails Environment
Write-Host "Checking Wails environment..." -ForegroundColor Cyan
try {
    $wailsVersion = wails version 2>$null
    if (-not $wailsVersion) {
        Write-Host "Error: Wails not found, please install Wails (go install github.com/wailsapp/wails/v2/cmd/wails@latest)" -ForegroundColor Red
        exit 1
    }
    Write-Host "Wails environment check passed" -ForegroundColor Green
} catch {
    Write-Host "Error: Wails environment check failed" -ForegroundColor Red
    exit 1
}

# Read and Encrypt .env Configuration
Write-Host "Reading .env configuration for embedding..." -ForegroundColor Cyan

# Check if .env exists in root
if (Test-Path ".env") {
    Write-Host "Found root .env file" -ForegroundColor Green
    
    # Check config directory
    if (-not (Test-Path "config")) {
        New-Item -ItemType Directory -Path "config" -Force | Out-Null
    }
    
    # Copy .env to config directory for go:embed
    Copy-Item ".env" -Destination "config\.env" -Force
    Write-Host "Copied .env to config/ directory for embedding" -ForegroundColor Green
} else {
    Write-Host "Warning: .env file not found in root directory!" -ForegroundColor Yellow
}

# Clean and Create dist Directory
Write-Host "Preparing dist directory..." -ForegroundColor Cyan
if (Test-Path "dist") {
    Remove-Item "dist" -Recurse -Force
    Write-Host "Old dist directory cleaned" -ForegroundColor Green
}

New-Item -ItemType Directory -Path "dist" -Force | Out-Null
Write-Host "dist directory created" -ForegroundColor Green

# Switch to wails directory for build
Write-Host "Switching to wails directory..." -ForegroundColor Cyan
Push-Location wails

# Build Binary
Write-Host "Building Wails application..." -ForegroundColor Cyan
try {
    # Execute wails build with injected ldflags
    
    # Default flags plus our injected flags
    $finalFlags = "-s -w"
    
    Write-Host "Build Flags: $finalFlags" -ForegroundColor Gray
    
    wails build -ldflags $finalFlags
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: Build failed" -ForegroundColor Red
        Pop-Location
        exit 1
    }
    
    # Move build artifact to root dist folder
    if (Test-Path "build\bin\wails.exe") {
        Copy-Item "build\bin\wails.exe" -Destination "..\dist\bilibili-tts-chat.exe"
        $fileSize = (Get-Item "..\dist\bilibili-tts-chat.exe").Length
        $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
        Write-Host "Build Success: bilibili-tts-chat.exe ($fileSizeMB MB)" -ForegroundColor Green
    } else {
        Write-Host "Error: Build completed but executable not found" -ForegroundColor Red
        Pop-Location
        exit 1
    }
} catch {
    Write-Host "Error: Build process failed: $_" -ForegroundColor Red
    Pop-Location
    exit 1
}

# Restore working directory
Pop-Location

# Copy Configuration Files
Write-Host "Copying and preparing configuration files..." -ForegroundColor Cyan

$configFiles = @(
    # .env is not copied because it's injected into the binary
    @{Source = "user.example.json"; Target = "user.json"}
    @{Source = "voices.json"; Target = "voices.json"}
)

foreach ($config in $configFiles) {
    $sourceFile = $config.Source
    $targetFile = $config.Target
    
    if (Test-Path $sourceFile) {
        try {
            if ($sourceFile -eq "user.example.json") {
                # Copy .example file to dist directory
                Copy-Item $sourceFile -Destination "dist\"
                Write-Host "Copied: $sourceFile to dist/" -ForegroundColor Green
                
                # Rename file in dist directory to remove .example suffix
                $distExampleFile = "dist\$sourceFile"
                $distTargetFile = "dist\$targetFile"
                
                if (Test-Path $distExampleFile) {
                    Rename-Item $distExampleFile -NewName $targetFile
                    Write-Host "Renamed: $sourceFile -> $targetFile" -ForegroundColor Green
                }
            } else {
                # Directly copy other config files
                Copy-Item $sourceFile -Destination "dist\"
                Write-Host "Copied: $sourceFile to dist/" -ForegroundColor Green
            }
        } catch {
            Write-Host "Error: Failed to process $sourceFile : $_" -ForegroundColor Red
        }
    } else {
        Write-Host "Warning: Source file not found: $sourceFile" -ForegroundColor Yellow
    }
}

# Show Results
Write-Host ""
Write-Host "=== Build Completed ===" -ForegroundColor Green
Write-Host "Build Directory: $(Get-Location)\dist" -ForegroundColor White
Write-Host ""
Write-Host "Files in dist directory:" -ForegroundColor Yellow
Get-ChildItem "dist" | ForEach-Object {
    $size = if ($_.PSIsContainer) { "Directory" } else { "$([math]::Round($_.Length / 1KB, 1))KB" }
    Write-Host "  $($_.Name) ($size)" -ForegroundColor White
}

Write-Host ""
Write-Host "Usage:" -ForegroundColor Yellow
Write-Host "  1. Run: .\bilibili-tts-chat.exe" -ForegroundColor White
Write-Host "  2. Configure Room ID Code in the interface on first run" -ForegroundColor White
Write-Host "  Note: API Keys are embedded, no .env file needed" -ForegroundColor Gray
