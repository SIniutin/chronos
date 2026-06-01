[CmdletBinding()]
param(
    [int]$WebPort = 3000,
    [string]$ApiBaseUrl = "http://localhost:8080",
    [switch]$SkipSeed
)

$ErrorActionPreference = "Stop"

$projectDir = $PSScriptRoot
$backendDir = Join-Path $projectDir "backend"
$frontendDir = Join-Path $projectDir "frontend"
$composeFile = Join-Path $backendDir "docker-compose.yml"
$databaseUrl = "postgres://postgres:postgres@localhost:5432/history-db?sslmode=disable"
$composeCommand = $null
$composeArgs = @()

function Assert-Command {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,

        [Parameter(Mandatory = $true)]
        [string]$InstallHint
    )

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command '$Name' was not found. $InstallHint"
    }
}

function Invoke-Checked {
    param(
        [Parameter(Mandatory = $true)]
        [string]$FilePath,

        [Parameter(Mandatory = $true)]
        [string[]]$ArgumentList
    )

    & $FilePath @ArgumentList
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed with exit code ${LASTEXITCODE}: $FilePath $($ArgumentList -join ' ')"
    }
}

function Invoke-Compose {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$ArgumentList
    )

    Invoke-Checked -FilePath $composeCommand -ArgumentList ($composeArgs + $ArgumentList)
}

function Test-NativeCommand {
    param(
        [Parameter(Mandatory = $true)]
        [string]$FilePath,

        [Parameter(Mandatory = $true)]
        [string[]]$ArgumentList
    )

    $previousErrorActionPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = "Continue"
        & $FilePath @ArgumentList *> $null
        return $LASTEXITCODE -eq 0
    }
    catch {
        return $false
    }
    finally {
        $ErrorActionPreference = $previousErrorActionPreference
    }
}

function Find-DockerComposeExecutable {
    $dockerCommand = Get-Command "docker" -ErrorAction Stop
    $dockerDir = Split-Path $dockerCommand.Source -Parent
    $dockerResourcesDir = Split-Path $dockerDir -Parent
    $candidates = @(
        (Join-Path $dockerResourcesDir "cli-plugins\docker-compose.exe"),
        (Join-Path $env:USERPROFILE ".docker\cli-plugins\docker-compose.exe"),
        (Join-Path $env:ProgramFiles "Docker\cli-plugins\docker-compose.exe")
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate -PathType Leaf) {
            return $candidate
        }
    }

    return $null
}

function Find-DockerDesktopExecutable {
    $candidates = @(
        (Join-Path $env:ProgramFiles "Docker\Docker\Docker Desktop.exe"),
        (Join-Path $env:LOCALAPPDATA "Programs\DockerDesktop\Docker Desktop.exe"),
        (Join-Path $env:LOCALAPPDATA "Docker\Docker Desktop.exe")
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate -PathType Leaf) {
            return $candidate
        }
    }

    return $null
}

function Test-DockerDesktopInstalled {
    return (Test-Path "HKLM:\SOFTWARE\Docker Inc.\Docker Desktop") -or
        (Test-Path "HKCU:\SOFTWARE\Docker Inc.\Docker Desktop")
}

function Initialize-DockerCompose {
    $composeExecutable = Find-DockerComposeExecutable
    if (($null -ne $composeExecutable) -and (Test-NativeCommand -FilePath $composeExecutable -ArgumentList @("version"))) {
        $script:composeCommand = $composeExecutable
        $script:composeArgs = @("-f", $composeFile, "--project-directory", $backendDir)
        Write-Host "Using Docker Compose executable: $composeExecutable"
        return
    }

    if (Test-NativeCommand -FilePath "docker" -ArgumentList @("compose", "version")) {
        $script:composeCommand = "docker"
        $script:composeArgs = @("compose", "-f", $composeFile, "--project-directory", $backendDir)
        Write-Host "Using Docker Compose v2 plugin."
        return
    }

    throw "Docker Compose was not found. Install the Docker Compose plugin or enable it in Docker Desktop."
}

function Test-WslAvailable {
    if (-not (Get-Command "wsl.exe" -ErrorAction SilentlyContinue)) {
        return $false
    }

    return (Test-NativeCommand -FilePath "wsl.exe" -ArgumentList @("--status")) -or
        (Test-NativeCommand -FilePath "wsl.exe" -ArgumentList @("--version"))
}

function Wait-ForDockerEngine {
    if (Test-NativeCommand -FilePath "docker" -ArgumentList @("info")) {
        Write-Host "Docker Engine is ready."
        return
    }

    if (-not (Test-DockerDesktopInstalled)) {
        throw "Docker Desktop installation is incomplete. Reinstall Docker Desktop, start it once, wait until Docker Engine is running, then run this script again."
    }

    if (-not (Test-WslAvailable)) {
        throw "Docker Engine is not running and WSL 2 is not installed. Open PowerShell as Administrator, run 'wsl --install', restart Windows, then reinstall or start Docker Desktop and run this script again."
    }

    $dockerDesktopExecutable = Find-DockerDesktopExecutable
    if ($null -eq $dockerDesktopExecutable) {
        throw "Docker Engine is not running. Start Docker Desktop and run this script again."
    }

    Write-Host "Docker Engine is not running. Starting Docker Desktop ..."
    Start-Process -FilePath $dockerDesktopExecutable -WindowStyle Hidden | Out-Null

    Write-Host "Waiting for Docker Engine ..."
    for ($attempt = 1; $attempt -le 120; $attempt++) {
        if (Test-NativeCommand -FilePath "docker" -ArgumentList @("info")) {
            Write-Host "Docker Engine is ready."
            return
        }

        Start-Sleep -Seconds 1
    }

    throw "Docker Desktop was started, but Docker Engine did not become ready within 120 seconds. Open Docker Desktop and review its status."
}

function Invoke-GoCommand {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Package
    )

    $previousDatabaseUrl = $env:DATABASE_URL
    try {
        $env:DATABASE_URL = $databaseUrl
        Push-Location $backendDir
        try {
            Invoke-Checked -FilePath "go" -ArgumentList @("run", $Package)
        }
        finally {
            Pop-Location
        }
    }
    finally {
        $env:DATABASE_URL = $previousDatabaseUrl
    }
}

function Wait-ForBackend {
    $healthUrl = "$($ApiBaseUrl.TrimEnd('/'))/health"
    Write-Host "Waiting for backend at $healthUrl ..."

    for ($attempt = 1; $attempt -le 60; $attempt++) {
        try {
            Invoke-RestMethod -Uri $healthUrl -Method Get -TimeoutSec 2 | Out-Null
            Write-Host "Backend is ready."
            return
        }
        catch {
            Start-Sleep -Seconds 1
        }
    }

    Write-Host "Backend logs:"
    Invoke-Compose -ArgumentList @("logs", "--tail", "100", "app")
    throw "Backend did not become ready within 60 seconds."
}

try {
    Assert-Command -Name "docker" -InstallHint "Install Docker Desktop and start it."
    Assert-Command -Name "go" -InstallHint "Install Go and add it to PATH."
    Assert-Command -Name "flutter" -InstallHint "Install Flutter and add its bin directory to PATH."

    Write-Host "Checking Docker Compose ..."
    Initialize-DockerCompose

    Wait-ForDockerEngine

    Write-Host "Starting PostgreSQL ..."
    Invoke-Compose -ArgumentList @("up", "-d", "db")
    Invoke-Compose -ArgumentList @(
        "exec",
        "-T",
        "db",
        "sh",
        "-c",
        "until pg_isready -U postgres -d history-db; do sleep 1; done"
    )

    Write-Host "Applying database migrations ..."
    Invoke-GoCommand -Package "./cmd/history-migrate"

    Write-Host "Building and starting backend services ..."
    Invoke-Compose -ArgumentList @("up", "-d", "--build", "app", "minio-init")
    Wait-ForBackend

    if (-not $SkipSeed) {
        Write-Host "Loading course seed data ..."
        Invoke-GoCommand -Package "./cmd/history-seed"
    }

    Write-Host "Installing Flutter dependencies ..."
    Push-Location $frontendDir
    try {
        Invoke-Checked -FilePath "flutter" -ArgumentList @("pub", "get")

        Write-Host ""
        Write-Host "Frontend URL: http://localhost:$WebPort"
        Write-Host "Backend URL:  $ApiBaseUrl"
        Write-Host "Press Ctrl+C to stop the Flutter web server."
        Write-Host "Backend containers will keep running for the next start."
        Write-Host ""

        Invoke-Checked -FilePath "flutter" -ArgumentList @(
            "run",
            "-d",
            "web-server",
            "--web-hostname",
            "0.0.0.0",
            "--web-port",
            "$WebPort",
            "--dart-define=API_BASE_URL=$ApiBaseUrl"
        )
    }
    finally {
        Pop-Location
    }
}
catch {
    Write-Error $_
    exit 1
}
