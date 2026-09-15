#/bin/sh

# Write your own script to delete dns records

echo "Some script that deletes dns records"

environment=$(printenv)
echo "Environment: "
echo "$environment"

echo ""

echo "Available Params:"
echo "Record (1): " "$1"
echo "Value (2): " "$2"
