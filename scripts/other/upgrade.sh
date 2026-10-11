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
else
  # The following segment was obtained from https://mywiki.wooledge.org/BashFAQ/062
  # Generate a temp directory to download the release.
  tempdir=""
  cleanup() { test "$tempdir" && rm -rf -- "$tempdir"; }
  trap cleanup EXIT HUP INT TERM    # makes so the function runs even if the script is killed
  tempdir=$(mktemp -d -- "${TMPDIR:-/tmp}/XXXXXXXXXXXXXX") ||
    { echo 'Error creating a temporary directory' >&2; exit 1; }
  cd "$tempdir" || exit 1

  # Get release link
  release_link="$(curl -s https://api.github.com/repos/gregtwallace/certwarden/releases\?per_page\=1 | grep -E '"html_url": ".*releases\/tag.*",' | grep -E '"https:\/\/.*"' -o)"

  # next lines from https://askubuntu.com/questions/89995/bash-remove-first-and-last-characters-from-a-string
  release_link="$(echo "$release_link" | cut -d '"' -f 2)"
  release_file="certwarden_linux_$(dpkg --print-architecture)"

  # Download the release file
  wget "$release_link/$release_file.zip"

  # Export the release
  unzip "$release_file.zip"
  tar -xf "$release_file.tar"
  cd "release/scripts" || exit 1

  # Update files.
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
