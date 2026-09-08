#Requires -Version 5.1
$ErrorActionPreference = 'Stop'

$BaseUrl = 'https://releases.codebahn.net/cli'
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\codebahn' }

function Get-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        'AMD64' { return 'amd64' }
        'ARM64' { return 'arm64' }
        default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
    }
}

function Get-LatestVersion {
    $json = Invoke-RestMethod -Uri "$BaseUrl/latest.json"
    return $json.version
}

function Test-Checksum {
    param([string]$File, [string]$ChecksumFile, [string]$Name)

    $lines = Get-Content $ChecksumFile
    $expected = $null
    foreach ($line in $lines) {
        $parts = $line -split '\s+'
        if ($parts.Length -ge 2 -and ($parts[1] -eq $Name -or $parts[1] -eq "*$Name")) {
            $expected = $parts[0]
            break
        }
    }
    if (-not $expected) {
        throw "No checksum found for $Name"
    }

    $actual = (Get-FileHash -Path $File -Algorithm SHA256).Hash.ToLower()
    if ($actual -ne $expected) {
        throw "Checksum mismatch: expected $expected, got $actual"
    }
}

function Install-Codebahn {
    $arch = Get-Arch
    $binary = "codebahn-windows-${arch}.exe"

    Write-Host "Fetching latest version... " -NoNewline
    $version = Get-LatestVersion
    Write-Host "v$version"

    $tag = "v$version"
    $tagUrl = "$BaseUrl/$tag"
    $tempDir = Join-Path $env:TEMP 'codebahn-install'
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    $tempBinary = Join-Path $tempDir 'codebahn.exe'
    $tempChecksums = Join-Path $tempDir 'checksums.txt'

    try {
        Write-Host "Downloading $binary... " -NoNewline
        Invoke-WebRequest -Uri "$tagUrl/$binary" -OutFile $tempBinary -UseBasicParsing
        Write-Host 'done'

        Write-Host 'Verifying checksum... ' -NoNewline
        Invoke-WebRequest -Uri "$tagUrl/checksums.txt" -OutFile $tempChecksums -UseBasicParsing
        Test-Checksum -File $tempBinary -ChecksumFile $tempChecksums -Name $binary
        Write-Host 'ok'

        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        Move-Item -Path $tempBinary -Destination (Join-Path $InstallDir 'codebahn.exe') -Force

        Write-Host "Installed codebahn v$version to $InstallDir\codebahn.exe"

        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if ($userPath -notlike "*$InstallDir*") {
            Write-Host ''
            Write-Host "Warning: $InstallDir is not in your PATH."
            Write-Host "Add it with:  `$env:Path = `"$InstallDir;`$env:Path`""
            Write-Host "Or permanently: [Environment]::SetEnvironmentVariable('Path', `"$InstallDir;`$([Environment]::GetEnvironmentVariable('Path', 'User'))`", 'User')"
        }
    }
    finally {
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Install-Codebahn
