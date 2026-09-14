#!/usr/bin/env python3

# This script transforms the 'stock' acme.sh files into the expected format for Cert Warden.

# Usage:
#   python3 ./acme_sh_update.py [release tag or hash]
# e.g., for 3.1.2:
#   python3 ./acme_sh_update.py 40290ad42a37aba57eb246e507c11944a52c0f68

import argparse
import io
import os
from pathlib import Path
import shutil
import tempfile
import urllib.request
import zipfile


# don't change pathing without also updating: `pkg/challenges/providers/dns01acmesh/cmd.go`
path_out_acmesh_root = os.path.join("scripts", "other", "acme.sh")
path_out_acmesh_dnsapi = os.path.join(path_out_acmesh_root, "dnsapi_cw")

###
###

# code for each dnsapi script to load the main source file
source_script = """
ABS_CURR_PATH=$(dirname $(realpath "${BASH_SOURCE[0]}"))
SRC_FILE="${ABS_CURR_PATH}/../acme_src.sh"
. "${SRC_FILE}"

"""

###
### Main Script

# parse args
parser = argparse.ArgumentParser()
parser.add_argument('tag')
args = parser.parse_args()

# download release
r = urllib.request.urlopen(f"https://github.com/acmesh-official/acme.sh/archive/{args.tag}.zip")
zf = zipfile.ZipFile(io.BytesIO(r.read()))

# extract releast to a temp location
with tempfile.TemporaryDirectory() as tempdir:
  zf.extractall(tempdir)


  ## Validation

  path_src_root = os.path.join(tempdir, f"acme.sh-{args.tag}")

  # path traversal check
  temp_as_path = Path(tempdir).resolve()
  path_src_root = Path(path_src_root).resolve()

  if not path_src_root.is_relative_to(temp_as_path):
    print("Security Error: Path traversal attempt detected.")
    exit(-5)

  # verify acme.sh main script exists
  path_src_acmesh_root_script = os.path.join(path_src_root, "acme.sh")
  if not os.path.isfile(path_src_acmesh_root_script):
    print("abort: acme.sh script vendor src not found")
    exit(-1)

  # verify license exists
  path_src_license = os.path.join(path_src_root, "LICENSE.md")
  if not os.path.isfile(os.path.join(path_src_license)):
    print("abort: acme.sh LICENSE.md vendor src not found")
    exit(-1)

  # verify acme.sh dnsapi folder exists
  path_src_dnsapi = os.path.join(path_src_root, "dnsapi")
  if not os.path.exists(path_src_dnsapi):
    print("abort: acme.sh dnsapi vendor src path not found")
    exit(-1)


  ## Nuke Old Files

  # delete previous version
  for item in os.listdir(path_out_acmesh_root):
    # keep the folder
    if item in [".gitkeep"]:
      continue

    # remove everything else
    item_abs = os.path.join(path_out_acmesh_root, item)

    if os.path.isdir(item_abs):
      shutil.rmtree(item_abs)
      continue

    os.remove(item_abs)


  ## Copy LICENSE.md
  shutil.copyfile(path_src_license, os.path.join(path_out_acmesh_root, "LICENSE.md"))


  ## Copy main script to use as source

  # read in main script
  acmeshData = ""
  with open(path_src_acmesh_root_script) as f:
    acmeshData = f.read()

  # remove line that runs main -- `main "$@"`
  acmeshData = acmeshData.replace('main "$@"', "")

  # write acme_src.sh
  acmeshSrcF = open(os.path.join(path_out_acmesh_root, "acme_src.sh"), "w")
  acmeshSrcF.write(acmeshData)


  ## Copy DNS APIs

  # create cw dns api folder
  if not os.path.exists(path_out_acmesh_dnsapi):
      os.makedirs(path_out_acmesh_dnsapi)

  # process each dnsapi file
  for filename in os.listdir(path_src_dnsapi):
    # only process scripts
    if not filename.endswith(".sh"):
      continue

    # read file in, preserve shebang, add source directive, and then the rest of the script
    dnsData = ""
    with open(os.path.join(path_src_dnsapi, filename), encoding="utf8") as f:
      # read in first line
      shebang = f.readline()

      # read the rest
      script = f.read()

      # combine
      dnsData = shebang + source_script + script

    # write to CW custom folder
    cwF = open(os.path.join(path_out_acmesh_dnsapi, filename), "w", encoding="utf8")
    cwF.write(dnsData)
