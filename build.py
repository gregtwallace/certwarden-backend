#!/usr/bin/env python3
import argparse
import os.path
from pathlib import Path
import re
import shutil
import subprocess

# Usage
# python3 ./build.py `target` [--gitrequired]

# `target` should be in the form `GOOS_GOARCH`.
# see: https://github.com/golang/go/blob/master/src/internal/syslist/syslist.go
# or unofficially: https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
# e.g.,
#  "windows_amd64",
#  "linux_amd64",
#  "linux_arm64",
#  "darwin_amd64",
#  "darwin_arm64",
#  "freebsd_amd64",
#  "freebsd_arm64",

# if the `--gitrequired` flag is used, the script will abort if it fails to get the current
# git commit


##
### Helper Functions
##

# build script is in the source root
PATH_SRC = os.path.dirname(os.path.realpath(__file__))

# return build path (useful for main build release script)
def output_path():
  return os.path.join(PATH_SRC, "_out")

# Function to get current commit hash without needing git executable or lib
# Modified version of: https://stackoverflow.com/a/68215738/7572076
def get_commit():
  git_folder = Path(os.path.join(PATH_SRC, '.git'))
  head_content = Path(git_folder, 'HEAD').read_text().split('\n')[0]
  commit_regex = re.compile(r"^[a-fA-F0-9]{40}$")

  # HEAD references another file in ref
  if head_content.startswith("ref: "):
    head_name = head_content.split(' ')[-1]
    head_ref = Path(git_folder,head_name)
    ref_file_content = head_ref.read_text().replace('\n','')

    if re.match(commit_regex, ref_file_content):
      return ref_file_content

  # HEAD has a commit in it (such as for a tag)
  if re.match(commit_regex, head_content):
    return head_content

  return ""

# get_backend_version parses the source code to get the backend version string
def get_backend_version():
  version_string = ""
  version_pattern = re.compile(r"appVersion = \"(\d+\.\d+\.\d+)\"")

  with open(os.path.join(PATH_SRC, 'pkg', 'domain', 'app', 'app.go')) as app_go_file:
    for line in app_go_file:
      match = re.search(version_pattern, line)
      if match != None:
        version_string = match.group(1)
        break

  if version_string == "":
    print("aborting: failed to parse backend version number")
    exit(-1)

  return version_string

# parse_os_target validates the os_target is properly formatted and returns the
# GOOS and GOARCH components. If the target is invalid, it will print an error and exit.
def parse_os_target(os_target):
  # validate
  if not re.match(r"^[a-zA-Z0-9]+_[a-zA-Z0-9]+$", os_target):
    print(f"aborting: invalid target '{os_target}'")
    exit(-1)

  # split and return
  split = os_target.split("_")
  goos = split[0]
  goarch = split[1]

  return goos, goarch

##
#### Build Script
##

def build(goos, goarch, git_required):
  # get version number
  version_string = get_backend_version()
  
  # try to get hash
  git_head = get_commit()
  if git_head != "":
    version_string += "_(" + git_head[:7] + ")"
  else:
    print("failed to get git hash")
    if git_required:
      print("aborting: git hash is required by --gitrequired")
      exit(-1)

  print(f"preparing to build certwarden-backend version '{version_string}' for target '{goos}_{goarch}'")

  os.environ["GOOS"] = goos
  os.environ["GOARCH"] = goarch
  os.environ["CGO_ENABLED"] = "0"

  # create out path
  path_output = output_path()
  if os.path.exists(path_output):
    print("certwarden-backend build output directory already exists, removing it")
    shutil.rmtree(path_output)
  os.makedirs(path_output)

  # special case for windows to add file extensions
  extension = ""
  if goos.lower() == "windows":
    extension = ".exe"

  # do build
  print("building certwarden-backend ...")

  # build binary
  result = subprocess.run(["go", "build", "-o", f"{path_output}/certwarden{extension}", "./cmd/api-server"], cwd=PATH_SRC)
  if result.returncode != 0:
    print("build certwarden-backend failed")
    exit(-2)

  # copy other important files for release
  shutil.copy(os.path.join(PATH_SRC, "config.default.yaml"), path_output)
  shutil.copy(os.path.join(PATH_SRC, "config.example.yaml"), path_output)
  shutil.copy(os.path.join(PATH_SRC, "config.changelog.md"), path_output)
  shutil.copy(os.path.join(PATH_SRC, "README.md"), path_output)
  shutil.copy(os.path.join(PATH_SRC, "LICENSE.md"), path_output)
  if git_head:
    with open(os.path.join(path_output, "HEAD-backend"), "a") as f:
      f.write(git_head)
  if goos.lower() == "windows":
    shutil.copytree(os.path.join(PATH_SRC, 'scripts', 'windows'), os.path.join(path_output, "scripts"))
  else:
    shutil.copytree(os.path.join(PATH_SRC, 'scripts', 'other'), os.path.join(path_output, "scripts"))

  print("... certwarden-backend build complete")

##
### Main
##

def main():
  print("initializing certwarden-backend build script")

  # parse args
  parser = argparse.ArgumentParser()
  parser.add_argument('target')
  parser.add_argument('--gitrequired', action='store_true')
  args = parser.parse_args()

  # build environment vars
  goos, goarch = parse_os_target(args.target)

  # do build
  build(goos, goarch, args.gitrequired)

  print("exiting certwarden-backend build script")

def temp():
  return os.path.dirname(os.path.realpath(__file__))

# run main if this script is called directly
if __name__ == "__main__":
  main()
