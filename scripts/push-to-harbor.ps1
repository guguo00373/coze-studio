# Push Coze Studio images to Harbor
# Usage: .\push-to-harbor.ps1 -Harbor "harbor.yourdomain.com"

param(
    [Parameter(Mandatory=$true)]
    [string]$Harbor,
    
    [string]$Project = "coze"
)

$images = @(
    "cozedev/coze-studio-server:latest"
    "cozedev/coze-studio-web:latest"
    "mysql:8.4.5"
    "bitnamilegacy/redis:8.0"
    "bitnamilegacy/elasticsearch:8.18.0"
    "minio/minio:RELEASE.2025-06-13T11-33-47Z-cpuv1"
    "milvusdb/milvus:v2.5.10"
    "bitnamilegacy/etcd:3.5"
    "nsqio/nsq:v1.2.1"
)

Write-Host "Logging in to $Harbor..." -ForegroundColor Cyan
docker login $Harbor

foreach ($img in $images) {
    $parts = $img -split "/"
    if ($parts.Count -gt 2) {
        $name = ($parts[1..($parts.Count-1)]) -join "/"
    } else {
        $name = $parts[1]
    }
    $tagged = "$Harbor/$Project/$name"
    
    Write-Host "Pushing $img -> $tagged" -ForegroundColor Yellow
    docker tag $img $tagged
    docker push $tagged
}

Write-Host "Done!" -ForegroundColor Green
