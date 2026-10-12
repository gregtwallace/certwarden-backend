#!/bin/sh

# install path and username
install_path="/opt/certwarden"
run_user="certwarden"


# Check for root
if [ "$(id -u)" -ne 0 ]; then echo "Please run as root." >&2; exit 1; fi

# move to script path
script_path=$(dirname $0)
cd "$script_path" || exit 1 # Fix for https://www.shellcheck.net/wiki/SC2164

# Check if using legacy method of updating (manual download of release)
if [ "$PWD" != "$install_path/scripts" ]; then
  # stop
  systemctl stop certwarden

  # copy new files
  rm -r "$install_path"/frontend_build
  cp -R ../* "$install_path"

  # permissions
  ./set_permissions.sh "$install_path" "$run_user"

  # allow binding to low port numbers
  setcap CAP_NET_BIND_SERVICE=+eip /opt/certwarden/certwarden

  # restart service
  systemctl start certwarden
else # Using the new method.
  # The following segment was obtained from https://mywiki.wooledge.org/BashFAQ/062
  # Generate a temp directory to download the release.
  tempdir=""
  cleanup() { test "$tempdir" && rm -rf -- "$tempdir"; }
  trap cleanup EXIT HUP INT TERM    # makes so the function runs even if the script is killed
  tempdir=$(mktemp -d -- "${TMPDIR:-/tmp}/tmpXXXXXXXXXXXXXX") ||
    { echo 'Error creating a temporary directory' >&2; exit 1; }
  cd "$tempdir" || exit 1
  # Get release link
  #The release_link line does the following:
  # 1. get the information about the lastest release of any kind. 2. Grab only the line with the release link.
  # 3. Remove unnessesary info (the variable name) 4. remove the quotes at the start and end.
  # 5. replace tag by download in the link.
  release_link="$(curl -s https://api.github.com/repos/gregtwallace/certwarden/releases\?per_page\=1 | grep -E '"html_url": ".*releases\/tag.*",' | grep -E '"https:\/\/.*"' -o | cut -d '"' -f 2 | sed "s/tag/download/")"
  # Determine what release asset is needed. the command outputs the architecture such as "amd64" or "arm64"
  release_file="certwarden_linux_$(dpkg --print-architecture)"
  # Download the release file
  wget "$release_link/$release_file.zip"
  # Export the release
  unzip "$release_file.zip"
  tar -xf "$release_file.tar"
  cd "release/scripts" || exit 1

  # Update like the legacy method.
    # stop
  systemctl stop certwarden

  # copy new files
  rm -r "$install_path"/frontend_build
  cp -R ../* "$install_path"

  # permissions
  ./set_permissions.sh "$install_path" "$run_user"

  # allow binding to low port numbers
  setcap CAP_NET_BIND_SERVICE=+eip /opt/certwarden/certwarden

  # restart service
  systemctl start certwarden
fi
