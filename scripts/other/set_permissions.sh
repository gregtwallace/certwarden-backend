#/bin/sh

# Default
default_path="/opt/certwarden"
default_user="certwarden"


# allow command line specified
install_path="${1:-$default_path}"
run_user="${2:-$default_user}"

# own entire folders / files
chown "$run_user":"$run_user" "$install_path" -R

# set to user/owner read/edit only but allow directory browsing
find "$install_path" -type d -exec chmod 755 {} \;
find "$install_path" -type f -exec chmod 640 {} \;

# main executable
chmod 750 "$install_path"/certwarden

# allow execution of scripts
find "$install_path"/scripts -type f -name "*.sh" -exec chmod 750 {} \;
