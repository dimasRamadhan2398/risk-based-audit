# Auto Launch All Services for AuditSphere (Frontend + Python AI Service)
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  Starting AuditSphere All-in-One Services..." -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Start Python AI Microservice (Port 8000)
Write-Host "[1/2] Launching Python AI Service on Port 8000..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PSScriptRoot\backend\python-ai'; python main.py"

# 2. Start Frontend Nuxt (Port 3000)
Write-Host "[2/2] Launching Frontend Nuxt Application on Port 3000..." -ForegroundColor Green
Set-Location -Path "$PSScriptRoot\frontend"
npm run dev
