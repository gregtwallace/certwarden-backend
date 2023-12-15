# Write your own script

Write-Host "Some script that post processes after order is valid."

# These env variables are always available, in addition to any custom ones
Write-Host "Environment LEGO_PRIVATE_KEY_NAME: $env:LEGO_PRIVATE_KEY_NAME"
Write-Host "Environment LEGO_PRIVATE_KEY_PEM: [redacted]" # $env:LEGO_PRIVATE_KEY_PEM
Write-Host "Environment LEGO_CERTIFICATE_NAME: $env:LEGO_CERTIFICATE_NAME"
Write-Host "Environment LEGO_CERTIFICATE_PEM: $env:LEGO_CERTIFICATE_PEM"
Write-Host "Environment LEGO_CERTIFICATE_COMMON_NAME: $env:LEGO_CERTIFICATE_COMMON_NAME"
