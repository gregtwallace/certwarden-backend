# Write your own script to create dns records

Write-Host "Some script that creates dns records"
Write-Host "Available Params:"

# WARNING: Domain is Deprecated.
# Domain will only return the last two pieces of the domain, so more
# complex domains will be truncated. For example something.in.ua would produce
# "in.ua" for this value, instead of "something.in.ua" !
Write-Host "Domain (args[0]): " $args[0] 

Write-Host "Record (args[1]): " $args[1] 
Write-Host "Value (args[2]): " $args[2]
