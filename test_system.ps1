# test_system.ps1
# Automated test script for Zcrypt log chain system

Write-Host "🔐 Zcrypt System Test Script" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$SERVER_URL = "http://localhost:8080"
$AGENT_ID = "test-agent-$(Get-Random -Maximum 9999)"
$AGENT_NAME = "Test Agent"

Write-Host "📋 Configuration:" -ForegroundColor Yellow
Write-Host "  Server URL: $SERVER_URL"
Write-Host "  Agent ID: $AGENT_ID"
Write-Host ""

# Step 1: Build the binaries
Write-Host "🔨 Step 1: Building binaries..." -ForegroundColor Green
go build -o zcrypt.exe ./agent
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to build client" -ForegroundColor Red
    exit 1
}
Write-Host "  ✓ Client built successfully" -ForegroundColor Green

go build -o zcrypt-server.exe ./server
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to build server" -ForegroundColor Red
    exit 1
}
Write-Host "  ✓ Server built successfully" -ForegroundColor Green
Write-Host ""

# Step 2: Generate keys
Write-Host "🔑 Step 2: Generating keypair..." -ForegroundColor Green
.\zcrypt.exe genkey
Write-Host ""

# Step 3: Start server in background
Write-Host "🚀 Step 3: Starting server..." -ForegroundColor Green
$serverJob = Start-Job -ScriptBlock {
    Set-Location $using:PWD
    .\zcrypt-server.exe
}
Write-Host "  ✓ Server started (Job ID: $($serverJob.Id))" -ForegroundColor Green
Start-Sleep -Seconds 2
Write-Host ""

# Step 4: Health check
Write-Host "💓 Step 4: Checking server health..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$SERVER_URL/api/v1/health" -Method Get
    Write-Host "  ✓ Server is healthy" -ForegroundColor Green
    Write-Host "    Status: $($response.status)" -ForegroundColor Gray
} catch {
    Write-Host "  ❌ Server health check failed" -ForegroundColor Red
    Stop-Job $serverJob
    Remove-Job $serverJob
    exit 1
}
Write-Host ""

# Step 5: Register agent
Write-Host "📝 Step 5: Registering agent..." -ForegroundColor Green
.\zcrypt.exe register-agent $AGENT_ID $AGENT_NAME
Write-Host ""

# Step 6: Send test logs
Write-Host "📤 Step 6: Sending test logs..." -ForegroundColor Green
$messages = @(
    "System startup",
    "User authentication successful",
    "File encryption: document.pdf",
    "Backup completed",
    "Security scan: no threats detected"
)

foreach ($msg in $messages) {
    Write-Host "  Sending: $msg" -ForegroundColor Gray
    .\zcrypt.exe send-to-server $msg
}
Write-Host ""

# Step 7: Verify local chain
Write-Host "✅ Step 7: Verifying local chain..." -ForegroundColor Green
.\zcrypt.exe chain-verify
Write-Host ""

# Step 8: Get local stats
Write-Host "📊 Step 8: Local chain statistics..." -ForegroundColor Green
.\zcrypt.exe chain-stats
Write-Host ""

# Step 9: Verify server chain
Write-Host "✅ Step 9: Verifying server chain..." -ForegroundColor Green
.\zcrypt.exe server-verify
Write-Host ""

# Step 10: Get server stats
Write-Host "📊 Step 10: Server statistics..." -ForegroundColor Green
.\zcrypt.exe server-stats
Write-Host ""

# Step 11: Test API directly
Write-Host "🔌 Step 11: Testing API endpoints..." -ForegroundColor Green

try {
    # Get all logs
    $logs = Invoke-RestMethod -Uri "$SERVER_URL/api/v1/logs?limit=5" -Method Get
    Write-Host "  ✓ GET /api/v1/logs - Returned $($logs.total) logs" -ForegroundColor Green
    
    # Get stats
    $stats = Invoke-RestMethod -Uri "$SERVER_URL/api/v1/stats" -Method Get
    Write-Host "  ✓ GET /api/v1/stats - $($stats.total_entries) entries" -ForegroundColor Green
    
    # Verify chain
    $verify = Invoke-RestMethod -Uri "$SERVER_URL/api/v1/verify/chain" -Method Post
    Write-Host "  ✓ POST /api/v1/verify/chain - Valid: $($verify.valid)" -ForegroundColor Green
} catch {
    Write-Host "  ❌ API test failed: $_" -ForegroundColor Red
}
Write-Host ""

# Step 12: Export chains
Write-Host "💾 Step 12: Exporting chains..." -ForegroundColor Green
.\zcrypt.exe chain-export
Write-Host "  ✓ Local chain exported to: zcrypt_chain_export.json" -ForegroundColor Green
Copy-Item server_logs.chain server_chain_export.json -ErrorAction SilentlyContinue
Write-Host "  ✓ Server chain copied to: server_chain_export.json" -ForegroundColor Green
Write-Host ""

# Cleanup
Write-Host "🧹 Cleaning up..." -ForegroundColor Yellow
Stop-Job $serverJob
Remove-Job $serverJob
Write-Host "  ✓ Server stopped" -ForegroundColor Green
Write-Host ""

# Summary
Write-Host "================================" -ForegroundColor Cyan
Write-Host "✨ Test Complete!" -ForegroundColor Green
Write-Host ""
Write-Host "📁 Generated Files:" -ForegroundColor Yellow
Write-Host "  - zcrypt.exe (CLI client)"
Write-Host "  - zcrypt-server.exe (Server)"
Write-Host "  - zcrypt_private.key (Your private key)"
Write-Host "  - zcrypt_public.key (Your public key)"
Write-Host "  - ~/.zcrypt/logs.chain (Local chain)"
Write-Host "  - server_logs.chain (Server chain)"
Write-Host "  - zcrypt_chain_export.json (Local export)"
Write-Host "  - server_chain_export.json (Server export)"
Write-Host ""
Write-Host "🎯 Next Steps:" -ForegroundColor Yellow
Write-Host "  1. Start server manually: .\zcrypt-server.exe"
Write-Host "  2. Send logs: .\zcrypt.exe send-to-server 'message'"
Write-Host "  3. Check stats: .\zcrypt.exe server-stats"
Write-Host ""
Write-Host "📖 For PostgreSQL integration (Step 3), see DATABASE_SETUP.md" -ForegroundColor Cyan