#!/usr/bin/env sh

ABS_CURR_PATH=$(dirname $(realpath "${BASH_SOURCE[0]}"))
SRC_FILE="${ABS_CURR_PATH}/../acme_src.sh"
. "${SRC_FILE}"

# shellcheck disable=SC2034
dns_knot_info='Knot Server knsupdate
Site: www.knot-dns.cz/docs/2.5/html/man_knsupdate.html
Docs: github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_knot
Options:
 KNOT_SERVER Server hostname. Default: "localhost".
 KNOT_KEY File path to TSIG key
'

# See also dns_nsupdate.sh

########  Public functions #####################

#Usage: dns_knot_add   _acme-challenge.www.domain.com   "XKrxpRBosdIKFzxW_CT3KLZNf6q0HG9i01zxXp5CPBs"
dns_knot_add() {
  fulldomain=$1
  txtvalue=$2
  _checkKey || return 1
  [ -n "${KNOT_SERVER}" ] || KNOT_SERVER="localhost"
  # save the dns server and key to the account.conf file.
  _saveaccountconf KNOT_SERVER "${KNOT_SERVER}"
  _saveaccountconf KNOT_KEY "${KNOT_KEY}"

  if ! _get_root "$fulldomain"; then
    _err "Domain does not exist."
    return 1
  fi

  _info "Adding ${fulldomain}. 60 TXT \"${txtvalue}\""

  knsupdate <<EOF
server ${KNOT_SERVER}
key ${KNOT_KEY}
zone ${_domain}.
update add ${fulldomain}. 60 TXT "${txtvalue}"
send
quit
EOF

  if [ $? -ne 0 ]; then
    _err "Error updating domain."
    return 1
  fi

  _info "Domain TXT record successfully added."
  return 0
}

#Usage: dns_knot_rm   _acme-challenge.www.domain.com
dns_knot_rm() {
  fulldomain=$1
  _checkKey || return 1
  [ -n "${KNOT_SERVER}" ] || KNOT_SERVER="localhost"

  if ! _get_root "$fulldomain"; then
    _err "Domain does not exist."
    return 1
  fi

  _info "Removing ${fulldomain}. TXT"

  knsupdate <<EOF
server ${KNOT_SERVER}
key ${KNOT_KEY}
zone ${_domain}.
update del ${fulldomain}. TXT
send
quit
EOF

  if [ $? -ne 0 ]; then
    _err "error updating domain"
    return 1
  fi

  _info "Domain TXT record successfully deleted."
  return 0
}

####################  Private functions below ##################################
# _acme-challenge.www.domain.com
# returns
# _domain=domain.com
_get_root() {
  domain=$1
  i="$(echo "$fulldomain" | tr '.' ' ' | wc -w)"
  i=$(_math "$i" - 1)

  while true; do
    h=$(printf "%s" "$domain" | cut -d . -f "$i"-100)
    if [ -z "$h" ]; then
      return 1
    fi
    _domain="$h"
    return 0
  done
  _debug "$domain not found"
  return 1
}

_checkKey() {
  if [ -z "${KNOT_KEY}" ]; then
    _err "You must specify a TSIG key to authenticate the request."
    return 1
  fi
}
