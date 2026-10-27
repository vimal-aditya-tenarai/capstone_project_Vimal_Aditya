$services = @(
  @{ID="auth-service-1";    Name="auth-service";    Address="localhost"; Port=8083; Check=@{HTTP="http://localhost:8083/health"; Interval="10s"}},
  @{ID="product-service-1"; Name="product-service"; Address="localhost"; Port=8081; Check=@{HTTP="http://localhost:8081/health"; Interval="10s"}},
  @{ID="order-service-1";   Name="order-service";   Address="localhost"; Port=8082; Check=@{HTTP="http://localhost:8082/health"; Interval="10s"}}
)

foreach ($s in $services) {
  $body = $s | ConvertTo-Json -Depth 5
  Invoke-RestMethod -Method Put -Uri "http://localhost:8500/v1/agent/service/register" -ContentType "application/json" -Body $body
}

Write-Host "Services registered with Consul successfully (auth-service, product-service, order-service). Check http://localhost:8500"
