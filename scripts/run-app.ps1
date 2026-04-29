param(
    [string]$LogsPath = "",
    [string]$CardsDbPath = "data/cards.json",
    [string]$BackendAddr = "127.0.0.1:8080",
    [string]$LiveApiAddr = "127.0.0.1:8765",
    [switch]$RebuildGo,
    [switch]$RebuildClient
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$clientDir = Join-Path $repoRoot "apps/hs-analyzer-client"
$binDir = Join-Path $repoRoot "bin"
$runDir = Join-Path $repoRoot ".tmp/run"
$outputDir = Join-Path $repoRoot "output"
$binarySuffix = if ($env:OS -eq "Windows_NT") { ".exe" } else { "" }
$backendBinary = Join-Path $binDir ("hs-analyzer-be" + $binarySuffix)
$cliBinary = Join-Path $binDir ("hs-analyzer-cli" + $binarySuffix)
$distEntry = Join-Path $clientDir "dist/index.html"
$electronEntry = Join-Path $clientDir "dist-electron/main/index.js"
$backendUrl = "http://$BackendAddr"

function Resolve-RepoPath([string]$PathValue) {
    if ([string]::IsNullOrWhiteSpace($PathValue)) {
        return ""
    }

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }

    return [System.IO.Path]::GetFullPath((Join-Path $repoRoot $PathValue))
}

function Assert-Command([string]$Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command '$Name' was not found in PATH."
    }
}

function Resolve-CommandPath([string]$PrimaryName, [string]$WindowsFallbackName = "", [switch]$PreferWindowsFallback) {
    $candidates = @()
    $isWindows = $env:OS -eq "Windows_NT"

    if ($isWindows -and $PreferWindowsFallback -and -not [string]::IsNullOrWhiteSpace($WindowsFallbackName)) {
        $candidates += $WindowsFallbackName
    }
    if (-not [string]::IsNullOrWhiteSpace($PrimaryName)) {
        $candidates += $PrimaryName
    }
    if ($isWindows -and -not $PreferWindowsFallback -and -not [string]::IsNullOrWhiteSpace($WindowsFallbackName)) {
        $candidates += $WindowsFallbackName
    }

    foreach ($candidate in $candidates) {
        $command = Get-Command $candidate -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($null -ne $command) {
            if (-not [string]::IsNullOrWhiteSpace($command.Path)) {
                return $command.Path
            }
            return $command.Name
        }
    }

    $displayName = if ([string]::IsNullOrWhiteSpace($WindowsFallbackName)) {
        $PrimaryName
    } else {
        "$PrimaryName/$WindowsFallbackName"
    }
    throw "Required command '$displayName' was not found in PATH."
}

function Invoke-CheckedCommand([string]$FilePath, [string[]]$Arguments, [string]$WorkingDirectory) {
    Push-Location $WorkingDirectory
    try {
        & $FilePath @Arguments
        if ($LASTEXITCODE -ne 0) {
            throw "Command failed: $FilePath $($Arguments -join ' ')"
        }
    }
    finally {
        Pop-Location
    }
}

function Get-LatestWriteTimeUtc([string[]]$Paths) {
    $latest = [datetime]::MinValue

    foreach ($path in $Paths) {
        if ([string]::IsNullOrWhiteSpace($path) -or -not (Test-Path -LiteralPath $path)) {
            continue
        }

        $items = @(Get-Item -LiteralPath $path)
        if ($items[0].PSIsContainer) {
            $items = @(
                Get-ChildItem -LiteralPath $path -Recurse -File -ErrorAction SilentlyContinue |
                    Where-Object { $_.FullName -notmatch "\\node_modules\\|\\dist\\|\\dist-electron\\|\\coverage\\|\\\.vite\\|\\\.tmp\\" }
            )
        }

        foreach ($item in $items) {
            if ($item.LastWriteTimeUtc -gt $latest) {
                $latest = $item.LastWriteTimeUtc
            }
        }
    }

    return $latest
}

function Test-ArtifactOutdated([string[]]$SourcePaths, [string[]]$ArtifactPaths) {
    foreach ($artifactPath in $ArtifactPaths) {
        if ([string]::IsNullOrWhiteSpace($artifactPath) -or -not (Test-Path -LiteralPath $artifactPath -PathType Leaf)) {
            return $true
        }
    }

    $latestSource = Get-LatestWriteTimeUtc $SourcePaths
    $latestArtifact = Get-LatestWriteTimeUtc $ArtifactPaths

    if ($latestSource -eq [datetime]::MinValue -or $latestArtifact -eq [datetime]::MinValue) {
        return $true
    }

    return $latestSource -gt $latestArtifact
}

function Start-BackgroundProcess(
    [string]$Name,
    [string]$FilePath,
    [string[]]$Arguments,
    [string]$WorkingDirectory
) {
    $stdoutPath = Join-Path $runDir "$Name.stdout.log"
    $stderrPath = Join-Path $runDir "$Name.stderr.log"
    $safeArguments = @($Arguments | Where-Object { $null -ne $_ })

    Write-Host ("Starting {0}..." -f $Name)
    $startProcessParams = @{
        FilePath               = $FilePath
        WorkingDirectory       = $WorkingDirectory
        RedirectStandardOutput = $stdoutPath
        RedirectStandardError  = $stderrPath
        WindowStyle            = "Hidden"
        PassThru               = $true
    }
    if ($safeArguments.Count -gt 0) {
        $startProcessParams.ArgumentList = $safeArguments
    }
    $process = Start-Process @startProcessParams

    Start-Sleep -Seconds 2
    if ($process.HasExited) {
        throw ("{0} exited early. Check logs:`n  {1}`n  {2}" -f $Name, $stdoutPath, $stderrPath)
    }

    Write-Host ("{0} PID={1}" -f $Name, $process.Id)
    return $process
}

$goCommand = Resolve-CommandPath "go"
$npmCommand = Resolve-CommandPath "npm" "npm.cmd" -PreferWindowsFallback

if (-not (Test-Path -LiteralPath $clientDir)) {
    throw "Electron client directory was not found: $clientDir"
}

if (-not (Test-Path -LiteralPath (Join-Path $clientDir "node_modules"))) {
    throw "Client dependencies are missing. Run 'npm install' in $clientDir first."
}

$cardsDbFullPath = Resolve-RepoPath $CardsDbPath
if (-not (Test-Path -LiteralPath $cardsDbFullPath -PathType Leaf)) {
    throw "Cards database was not found: $cardsDbFullPath"
}

$logsFullPath = Resolve-RepoPath $LogsPath
if ($logsFullPath -and -not (Test-Path -LiteralPath $logsFullPath)) {
    throw "Logs path was not found: $logsFullPath"
}

New-Item -ItemType Directory -Force -Path $binDir | Out-Null
New-Item -ItemType Directory -Force -Path $runDir | Out-Null
New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $repoRoot ".tmp/gocache") | Out-Null

$originalGoCache = $env:GOCACHE
$env:GOCACHE = Join-Path $repoRoot ".tmp/gocache"

try {
    $backendSources = @(
        (Join-Path $repoRoot "apps/hs-analyzer-be"),
        (Join-Path $repoRoot "internal"),
        (Join-Path $repoRoot "pkg"),
        (Join-Path $repoRoot "go.mod"),
        (Join-Path $repoRoot "go.sum")
    )
    $cliSources = @(
        (Join-Path $repoRoot "apps/hs-analyzer-cli"),
        (Join-Path $repoRoot "internal"),
        (Join-Path $repoRoot "pkg"),
        (Join-Path $repoRoot "cmd"),
        (Join-Path $repoRoot "go.mod"),
        (Join-Path $repoRoot "go.sum")
    )
    $clientSources = @(
        (Join-Path $clientDir "electron"),
        (Join-Path $clientDir "src"),
        (Join-Path $clientDir "index.html"),
        (Join-Path $clientDir "package.json"),
        (Join-Path $clientDir "tsconfig.json"),
        (Join-Path $clientDir "tsconfig.electron.json"),
        (Join-Path $clientDir "vite.config.ts"),
        (Join-Path $clientDir "tailwind.config.ts"),
        (Join-Path $clientDir "postcss.config.cjs")
    )

    if ($RebuildGo -or (Test-ArtifactOutdated $backendSources @($backendBinary))) {
        Write-Host "Building backend binary..."
        Invoke-CheckedCommand $goCommand @("build", "-o", $backendBinary, "./apps/hs-analyzer-be/cmd/hs-analyzer-be") $repoRoot
    }

    if ($RebuildGo -or (Test-ArtifactOutdated $cliSources @($cliBinary))) {
        Write-Host "Building CLI binary..."
        Invoke-CheckedCommand $goCommand @("build", "-o", $cliBinary, "./apps/hs-analyzer-cli/cmd/hsanalyzer") $repoRoot
    }

    if ($RebuildClient -or (Test-ArtifactOutdated $clientSources @($distEntry, $electronEntry))) {
        Write-Host "Building Electron client..."
        Invoke-CheckedCommand $npmCommand @("run", "build:app") $clientDir
    }

    $processes = @()
    try {
        $processes += Start-BackgroundProcess "backend" $backendBinary @() $repoRoot

        $monitorArgs = @(
            "monitor",
            "--cards-db-path", $cardsDbFullPath,
            "--live-api-addr", $LiveApiAddr,
            "--match-upload-url", "$backendUrl/api/v1/matches"
        )
        if ($logsFullPath) {
            $monitorArgs += @("--logs-path", $logsFullPath)
        }

        $processes += Start-BackgroundProcess "monitor" $cliBinary $monitorArgs $repoRoot

        Write-Host ""
        Write-Host "Local services are ready:"
        Write-Host ("  backend: {0}" -f $backendUrl)
        Write-Host ("  live api: http://{0}" -f $LiveApiAddr)
        Write-Host ("  logs: {0}" -f $runDir)
        Write-Host ""

        Invoke-CheckedCommand $npmCommand @("run", "start") $clientDir
    }
    finally {
        foreach ($process in $processes) {
            if ($null -ne $process -and -not $process.HasExited) {
                Stop-Process -Id $process.Id -Force
            }
        }
    }
}
finally {
    $env:GOCACHE = $originalGoCache
}
