#/bin/sh

# Write your own script to create dns records

echo "Some script that creates dns records"
echo "Available Params:"

# WARNING: Domain is Deprecated.
# Domain will only return the last two pieces of the domain, so more
# complex domains will be truncated. For example something.in.ua would produce
# "in.ua" for this value, instead of "something.in.ua" !
echo "Domain (1): " "$1"

echo "Record (2): " "$2"
echo "Value (3): " "$3"
