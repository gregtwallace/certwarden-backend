#!/usr/bin/env python3

# This script transforms the 'stock' acme.sh files into the expected format
# for Cert Warden. This script should be run any time the acme.sh version
# is bumped.

# This shifts the script processing to once pre-release vs. repeatedly during
# runtime.

import os
import shutil

# keep in sync with `pkg\challenges\providers\dns01acmesh\cmd.go`
dnsApiCwPath = "/dnsapi_cw"

###
###

# make dnsapi path relative
dnsApiCwPath = "." + dnsApiCwPath

# verify acme.sh is in the current path
if not os.path.exists("acme.sh"):
  print("abort: acme.sh not found in current working directory")
  exit(-1)

if not os.path.exists("dnsapi"):
  print("abort: acme.sh dnsapi path not found in current working directory")
  exit(-1)

# delete any previously generated files
if os.path.exists(dnsApiCwPath):
  shutil.rmtree(dnsApiCwPath)
if os.path.exists("acme_src.sh"):
  os.remove("acme_src.sh")

# read in main script
acmeshData = ""
with open('acme.sh') as f:
  acmeshData = f.read()

# remove line that runs main -- `main "$@"`
acmeshData = acmeshData.replace('main "$@"', "")

# write acme_src.sh
acmeshSrcF = open("acme_src.sh", "w")
acmeshSrcF.write(acmeshData)

# create cw folder if doesn't exist
if not os.path.exists(dnsApiCwPath):
    os.makedirs(dnsApiCwPath)

# process each dnsapi file
for filename in os.listdir("dnsapi"):
  # only process scripts
  if not filename.endswith(".sh"):
    continue

  # read file in, preserve shebang, add source directive, and then the rest of the script
  dnsData = ""
  with open("dnsapi/" + filename) as f:
    # read in first line
    shebang = f.readline()

    # read the rest
    script = f.read()

    # combine
    dnsData = shebang + ". ../acme_src.sh\n\n" + script

  # write to CW custom folder
  cwF = open(dnsApiCwPath + "/" + filename, "w")
  cwF.write(dnsData)
