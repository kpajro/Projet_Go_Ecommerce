<#
.SYNOPSIS
    Raccourcis de développement P_Go pour Windows.
.DESCRIPTION
    Équivalent du Makefile, car `make` n'est pas installé nativement sur Windows.
.EXAMPLE
    .\scripts\dev.ps1 run-server
    .\scripts\dev.ps1 check
#>
param(
    [ValidateSet('tidy', 'fmt', 'vet', 'build', 'test', 'run-server', 'run-client', 'run-admin', 'docker-up', 'docker-down', 'docker-logs', 'psql', 'check', 'help')]
    [string]$Task = 'help'
)

$ErrorActionPreference = 'Stop'
Set-Location (Split-Path -Parent $PSScriptRoot)

switch ($Task) {
    'tidy'        { go mod tidy }
    'fmt'         { go fmt ./... }
    'vet'         { go vet ./... }
    'build'       {
        go build -o bin/pgo-server.exe ./cmd/server
        go build -o bin/pgo-client.exe ./cmd/client
        go build -o bin/pgo-admin.exe ./cmd/admin
        Write-Host "Binaires produits dans bin\" -ForegroundColor Green
    }
    'test'        { go test ./... -count=1 }
    'run-server'  { go run ./cmd/server }
    'run-client'  { go run ./cmd/client }
    'run-admin'   { go run ./cmd/admin }
    'docker-up'   { docker compose up -d --build }
    'docker-down' { docker compose down }
    'docker-logs' { docker compose logs -f }
    'psql'        { docker compose exec postgres psql -U pgo -d pgo }
    'check'       {
        go fmt ./...
        go vet ./...
        go test ./... -count=1
    }
    default       { Get-Help $PSCommandPath -Detailed }
}
