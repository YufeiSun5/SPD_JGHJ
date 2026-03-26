# =============================================
# 手动触发配置热重载
# =============================================

# 配置 API 地址
$apiUrl = "http://localhost:8080/api/v1/config/reload"

Write-Host "正在触发配置热重载..." -ForegroundColor Yellow

try {
    $response = Invoke-RestMethod -Uri $apiUrl -Method POST -ContentType "application/json"
    
    Write-Host "✅ 配置重载成功!" -ForegroundColor Green
    Write-Host "测点数量: $($response.tags)" -ForegroundColor Cyan
    Write-Host "网关数量: $($response.gateways)" -ForegroundColor Cyan
    Write-Host "响应消息: $($response.message)" -ForegroundColor Cyan
}
catch {
    Write-Host "❌ 配置重载失败!" -ForegroundColor Red
    Write-Host "错误信息: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "服务器响应: $responseBody" -ForegroundColor Red
    }
}

# =============================================
# 查询当前配置版本
# =============================================
Write-Host "`n正在查询当前配置版本..." -ForegroundColor Yellow

try {
    $versionUrl = "http://localhost:8080/api/v1/config/version"
    $versionResponse = Invoke-RestMethod -Uri $versionUrl -Method GET
    
    Write-Host "✅ 当前配置版本: $($versionResponse.version)" -ForegroundColor Green
    Write-Host "时间戳: $($versionResponse.timestamp)" -ForegroundColor Cyan
}
catch {
    Write-Host "❌ 查询版本失败: $($_.Exception.Message)" -ForegroundColor Red
}

























