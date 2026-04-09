$ErrorActionPreference = "Stop"

$arch = if ([Environment]::Is64BitOperatingSystem) {
    if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) {
        "arm64"
    } else {
        "amd64"
    }
} else {
    Write-Error "Unsupported architecture"
    exit 1
}

$binary = "lazy-mcp-windows-${arch}.exe"
$url = "https://github.com/tkukushkin/lazy-mcp/releases/latest/download/${binary}"

$destDir = Join-Path $env:LOCALAPPDATA "lazy-mcp"
$dest = Join-Path $destDir "lazy-mcp.exe"

if (-not (Test-Path $destDir)) {
    New-Item -ItemType Directory -Path $destDir | Out-Null
}

Write-Host "Downloading lazy-mcp for windows/${arch}..."
Invoke-WebRequest -Uri $url -OutFile $dest -UseBasicParsing

# Add to PATH if not already there
if ($env:PATH -notlike "*$destDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$([Environment]::GetEnvironmentVariable("PATH", "User"));$destDir", "User")
    Write-Host "Added ${destDir} to user PATH (restart your terminal to apply)"
}

Write-Host "Installed lazy-mcp to ${dest}"
