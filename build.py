#!/usr/bin/env python3
import argparse
import os.path
from pathlib import Path
import re
import shutil
import subprocess
import tarfile

# Usage
# python3 ./build_release.py `target` [--gitrequired]

# target should be in the form `GOOS_GOARCH` (see below).

# if the gitrequired flag is used, the script will abort if it fails to get the current
# git commit

# Assumptions:
# Backend source is cloned into [root]/[src]/certwarden-backend

# target strings must be in the format:
#   `GOOS_GOARCH`
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

##
### Helper Functions
##

# Function to get current commit hash without needing git executable or lib
# Modified version of: https://stackoverflow.com/a/68215738/7572076
def get_commit(src_path):
  git_folder = Path(os.path.join(src_path, '.git'))
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


##
### Main Script
##

print("initializing certwarden-backend build script")

# define paths
path_src_backend = os.path.dirname(os.path.realpath(__file__))
path_src = Path(__file__).parents[1]
path_root = Path(__file__).parents[2]

path_output = os.path.join(path_root, "_out", "backend")

# parse args
parser = argparse.ArgumentParser()
parser.add_argument('target')
parser.add_argument('--gitrequired', action='store_true')
args = parser.parse_args()

# get version number
versionString = ""
versionPattern = re.compile(r"appVersion = \"(\d+\.\d+\.\d+)\"")

with open(os.path.join(path_src_backend, 'pkg', 'domain', 'app', 'app.go')) as appGoFile:
  for line in appGoFile:
    match = re.search(versionPattern, line)
    if match != None:
      versionString = match.group(1)
      break

if versionString == "":
  print("aborting: failed to parse version number")
  exit(-1)

# try to get hash
gitHead = get_commit(path_src_backend)
if gitHead != "":
  versionString += "_(" + gitHead[:7] + ")"
else:
  print("failed to get git hash")
  if args.gitrequired:
    print("aborting: git hash is required by --gitrequired")
    exit(-1)

#
print("building certwarden-backend version", versionString)

# recreate paths
if os.path.exists(path_output):
  print("build output directory already exists, removing it")
  shutil.rmtree(path_output)
os.makedirs(path_output)

# build target
target = args.target
print("building certwarden-backend for target:", target, "...")

# environment vars
split = target.split("_")
GOOS = split[0]
GOARCH = split[1]
os.environ["GOOS"] = GOOS
os.environ["GOARCH"] = GOARCH
os.environ["CGO_ENABLED"] = "0"

# send build product to GOOS_GOARCH subfolders
targetOutDir = os.path.join(path_output, target)
if not os.path.exists(targetOutDir):
  os.makedirs(targetOutDir)

# special case for windows to add file extensions
extension = ""
if GOOS.lower() == "windows":
  extension = ".exe"

# build binary
result = subprocess.run(["go", "build", "-o", f"{targetOutDir}/certwarden{extension}", "./cmd/api-server"], cwd=path_src_backend)
if result.returncode != 0:
  print(f"build certwarden-backend target '{target}' failed")
  exit(-2)

# copy other important files for release
shutil.copy(os.path.join(path_src_backend, "config.default.yaml"), targetOutDir)
shutil.copy(os.path.join(path_src_backend, "config.example.yaml"), targetOutDir)
shutil.copy(os.path.join(path_src_backend, "config.changelog.md"), targetOutDir)
shutil.copy(os.path.join(path_src_backend, "README.md"), targetOutDir)
shutil.copy(os.path.join(path_src_backend, "LICENSE.md"), targetOutDir)
if gitHead:
  with open(targetOutDir + "/HEAD-backend", "a") as f:
    f.write(gitHead)
if GOOS.lower() == "windows":
  shutil.copytree(os.path.join(path_src_backend, 'scripts', 'windows'), os.path.join(targetOutDir, "scripts"))
else:
  shutil.copytree(os.path.join(path_src_backend, 'scripts', 'other'), os.path.join(targetOutDir, "scripts"))

print("exiting certwarden-backend build script")
