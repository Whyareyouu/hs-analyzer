param(
    [string]$Locale = "ruRU",
    [string]$OutFile = "data/cards.json"
)

$ErrorActionPreference = "Stop"

$url = "https://api.hearthstonejson.com/v1/latest/$Locale/cards.json"
Write-Host "Downloading cards database from $url"

Invoke-WebRequest -Uri $url -OutFile $OutFile

$file = Get-Item $OutFile
Write-Host ("Saved: {0} ({1} bytes)" -f $file.FullName, $file.Length)
