# Write your own script to create dns records

Write-Host "Some script that creates dns records"

Write-Host "Environment: $(Get-ChildItem env: | Out-String)"

Write-Host "Available Params:"
Write-Host "Record (args[0]): " $args[0] 
Write-Host "Value (args[1]): " $args[1]
