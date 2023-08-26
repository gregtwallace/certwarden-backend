#/bin/sh

# Write your own script to create dns records

echo "Some script that creates dns records"

environment=$(printenv)
echo "Environment: "
echo "$environment"

echo ""

echo "Available Params:"
echo "Domain (1): " "$1"
echo "Record (2): " "$2"
echo "Value (3): " "$3"
