#/bin/sh

# Write your own script
echo "Some script that post processes"

# These env variables are always available, in addition to any custom ones
echo "Environment LEGO_PRIVATE_KEY_NAME: $(printenv LEGO_PRIVATE_KEY_NAME)"
echo "Environment LEGO_PRIVATE_KEY_PEM: [redacted]" # $(printenv LEGO_PRIVATE_KEY_PEM)
echo "Environment LEGO_CERTIFICATE_NAME: $(printenv LEGO_CERTIFICATE_NAME)"
echo "Environment LEGO_CERTIFICATE_PEM: $(printenv LEGO_CERTIFICATE_PEM)"
echo "Environment LEGO_CERTIFICATE_COMMON_NAME: $(printenv LEGO_CERTIFICATE_COMMON_NAME)"
