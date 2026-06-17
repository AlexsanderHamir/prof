# Run the full test suite and print statement coverage using the Go toolchain.
#
# Usage (from repo root):
#   .\scripts\test-cover.ps1
#   .\scripts\test-cover.ps1 -WriteHtml
param(
    [string]$Profile = "coverage.out",
    [string]$HtmlPath = "coverage.html",
    [switch]$WriteHtml
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root

Write-Host "Running tests with coverage (-coverpkg=./...)..."
$pkgs = go list ./... | Where-Object { $_ -notmatch '/scripts/' }
go test @pkgs `
    -count=1 `
    -covermode atomic `
    -coverpkg ./... `
    -coverprofile $Profile
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "=== Coverage summary ==="
$coverMain = Join-Path $Root "scripts\coverreport\main.go"
if ($WriteHtml) {
    go run $coverMain -html $HtmlPath $Profile
} else {
    go run $coverMain $Profile
}
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "Profile: $(Join-Path $Root $Profile)"
if ($WriteHtml) {
    Write-Host "HTML:    $(Join-Path $Root $HtmlPath)"
}
Write-Host "Tip: go tool cover -func $Profile"
