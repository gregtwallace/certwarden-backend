# Write your own script to delete dns records

Write-Host "Some script that deletes dns records"

Write-Host "Environment: $(Get-ChildItem env: | Out-String)"

Write-Host "Available Params:"
Write-Host "Record (args[0]): " $args[0] 
Write-Host "Value (args[1]): " $args[1]
