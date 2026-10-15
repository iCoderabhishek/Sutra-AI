# Installs the latest sutra CLI release for Windows.
#   irm https://raw.githubusercontent.com/iCoderabhishek/Sutra-AI/HEAD/scripts/install.ps1 | iex
# Set $env:SUTRA_INSTALL_DIR to change the install location.
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$repo = 'iCoderabhishek/Sutra-AI'
$installDir = if ($env:SUTRA_INSTALL_DIR) { $env:SUTRA_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'sutra\bin' }

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

# Older Windows PowerShell may default to TLS 1.0, which GitHub rejects.
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

# Download with visible progress and timeouts so a stalled connection fails
# instead of hanging forever.
#   connect: give up if the server doesn't answer in 15s
#   stall:   abort if speed stays under 1 KB/s for 30s
#   total:   abort after 5 minutes no matter what
function Get-File([string]$Url, [string]$OutFile) {
    $curl = Get-Command curl.exe -CommandType Application -ErrorAction SilentlyContinue
    if ($curl) {
        # curl.exe ships with Windows 10 1803+; its progress bar is accurate and cheap.
        & $curl.Source --fail --location --progress-bar --retry 2 `
            --connect-timeout 15 --speed-limit 1024 --speed-time 30 --max-time 300 `
            --user-agent 'sutra-installer' --output $OutFile $Url
        if ($LASTEXITCODE -ne 0) {
            throw "Download failed or timed out (curl exit $LASTEXITCODE): $Url"
        }
        return
    }

    # Fallback: show PowerShell's own progress bar for this download only.
    $old = $ProgressPreference
    $ProgressPreference = 'Continue'
    try {
        Invoke-WebRequest $Url -OutFile $OutFile -UseBasicParsing -TimeoutSec 300 `
            -Headers @{ 'User-Agent' = 'sutra-installer' }
    }
    catch {
        throw "Download failed or timed out: $Url ($($_.Exception.Message))"
    }
    finally {
        $ProgressPreference = $old
    }
}

$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest" `
    -Headers @{ 'User-Agent' = 'sutra-installer' } -UseBasicParsing -TimeoutSec 30
$tag = $release.tag_name
$version = $tag.TrimStart('v')
$archive = "sutra_${version}_windows_${arch}.zip"
$base = "https://github.com/$repo/releases/download/$tag"

$tmp = Join-Path ([IO.Path]::GetTempPath()) ("sutra-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    Write-Host "Downloading sutra $tag (windows/$arch)..."
    Get-File "$base/$archive" (Join-Path $tmp $archive)
    Get-File "$base/checksums.txt" (Join-Path $tmp 'checksums.txt')

    # Verify the archive against the published checksums.
    $line = Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { $_ -match " $([regex]::Escape($archive))$" }
    if (-not $line) { throw "No checksum found for $archive" }
    $expected = ($line -split '\s+')[0]
    $actual = (Get-FileHash (Join-Path $tmp $archive) -Algorithm SHA256).Hash
    if ($expected -ne $actual.ToLower()) { throw "Checksum mismatch for $archive" }

    Expand-Archive (Join-Path $tmp $archive) -DestinationPath $tmp -Force
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    Copy-Item (Join-Path $tmp 'sutra.exe') (Join-Path $installDir 'sutra.exe') -Force
}
finally {
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

# Add the install dir to the user PATH once.
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $installDir) {
    [Environment]::SetEnvironmentVariable('Path', "$installDir;$userPath", 'User')
    $env:Path = "$installDir;$env:Path"
    Write-Host "Added $installDir to your PATH. Open a new terminal if 'sutra' is not found."
}

Write-Host "Installed sutra $tag to $installDir\sutra.exe"
Write-Host "Run: sutra"
