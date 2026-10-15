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

$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest" -Headers @{ 'User-Agent' = 'sutra-installer' }
$tag = $release.tag_name
$version = $tag.TrimStart('v')
$archive = "sutra_${version}_windows_${arch}.zip"
$base = "https://github.com/$repo/releases/download/$tag"

$tmp = Join-Path ([IO.Path]::GetTempPath()) ("sutra-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    Write-Host "Downloading sutra $tag (windows/$arch)..."
    Invoke-WebRequest "$base/$archive" -OutFile (Join-Path $tmp $archive)
    Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt')

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
