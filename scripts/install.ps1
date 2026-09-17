#Requires -Version 5.1
$ErrorActionPreference = 'Stop'
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

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

function Test-Signature {
    param([string]$ChecksumFile, [string]$TagUrl)

    $sigFile = Join-Path (Split-Path $ChecksumFile) 'checksums.txt.asc'
    Invoke-WebRequest -Uri "$TagUrl/checksums.txt.asc" -OutFile $sigFile -UseBasicParsing

    $gpg = Get-Command gpg -ErrorAction SilentlyContinue
    if (-not $gpg) {
        Write-Host 'WARNING: gpg not found; PGP signature verification skipped.'
        Write-Host '         Install GPG for Windows and re-run, or verify manually:'
        Write-Host "           gpg --verify `"$sigFile`" `"$ChecksumFile`""
        return
    }

    $keyFile = Join-Path (Split-Path $ChecksumFile) 'release-key.asc'
    Invoke-WebRequest -Uri "$BaseUrl/release-key.asc" -OutFile $keyFile -UseBasicParsing
    & gpg --batch --import $keyFile 2>$null

    $null = & gpg --batch --verify $sigFile $ChecksumFile 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw 'PGP signature verification failed. The checksums file may have been tampered with.'
    }
    Write-Host 'ok'
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

        Write-Host 'Verifying PGP signature... ' -NoNewline
        Test-Signature -ChecksumFile $tempChecksums -TagUrl $tagUrl

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
