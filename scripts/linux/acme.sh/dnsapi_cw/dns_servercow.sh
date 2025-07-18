#!/usr/bin/env sh

ABS_CURR_PATH=$(dirname $(realpath "${BASH_SOURCE[0]}"))
SRC_FILE="${ABS_CURR_PATH}/../acme_src.sh"
. "${SRC_FILE}"

# shellcheck disable=SC2034
dns_servercow_info='ServerCow.de
Site: ServerCow.de
Docs: github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_servercow
Options:
 SERVERCOW_API_Username Username
 SERVERCOW_API_Password Password
Issues: github.com/jhartlep/servercow-dns-api/issues
Author: Jens Hartlep
'

SERVERCOW_API="https://api.servercow.de/dns/v1/domains"

# Usage dns_servercow_add _acme-challenge.www.domain.com "abcdefghijklmnopqrstuvwxyz"
dns_servercow_add() {
  fulldomain=$1
  txtvalue=$2

  _info "Using servercow"
  _debug fulldomain "$fulldomain"
  _debug txtvalue "$txtvalue"

  SERVERCOW_API_Username="${SERVERCOW_API_Username:-$(_readaccountconf_mutable SERVERCOW_API_Username)}"
  SERVERCOW_API_Password="${SERVERCOW_API_Password:-$(_readaccountconf_mutable SERVERCOW_API_Password)}"
  if [ -z "$SERVERCOW_API_Username" ] || [ -z "$SERVERCOW_API_Password" ]; then
    SERVERCOW_API_Username=""
    SERVERCOW_API_Password=""
    _err "You don't specify servercow api username and password yet."
    _err "Please create your username and password and try again."
    return 1
  fi

  # save the credentials to the account conf file
  _saveaccountconf_mutable SERVERCOW_API_Username "$SERVERCOW_API_Username"
  _saveaccountconf_mutable SERVERCOW_API_Password "$SERVERCOW_API_Password"

  _debug "First detect the root zone"
  if ! _get_root "$fulldomain"; then
    _err "invalid domain"
    return 1
  fi

  _debug _sub_domain "$_sub_domain"
  _debug _domain "$_domain"

  # check whether a txt record already exists for the subdomain
  if printf -- "%s" "$response" | grep "{\"name\":\"$_sub_domain\",\"ttl\":20,\"type\":\"TXT\"" >/dev/null; then
    _info "A txt record with the same name already exists."
    # trim the string on the left
    txtvalue_old=${response#*{\"name\":\""$_sub_domain"\",\"ttl\":20,\"type\":\"TXT\",\"content\":\"}
    # trim the string on the right
    txtvalue_old=${txtvalue_old%%\"*}

    _debug txtvalue_old "$txtvalue_old"

    _info "Add the new txtvalue to the existing txt record."
    if _servercow_api POST "$_domain" "{\"type\":\"TXT\",\"name\":\"$fulldomain\",\"content\":[\"$txtvalue\",\"$txtvalue_old\"],\"ttl\":20}"; then
      if printf -- "%s" "$response" | grep "ok" >/dev/null; then
        _info "Added additional txtvalue, OK"
        return 0
      else
        _err "add txt record error."
        return 1
      fi
    fi
    _err "add txt record error."
    return 1
  else
    _info "There is no txt record with the name yet."
    if _servercow_api POST "$_domain" "{\"type\":\"TXT\",\"name\":\"$fulldomain\",\"content\":\"$txtvalue\",\"ttl\":20}"; then
      if printf -- "%s" "$response" | grep "ok" >/dev/null; then
        _info "Added, OK"
        return 0
      else
        _err "add txt record error."
        return 1
      fi
    fi
    _err "add txt record error."
    return 1
  fi

}

# Usage fulldomain txtvalue
# Remove the txt record after validation
dns_servercow_rm() {
  fulldomain=$1
  txtvalue=$2

  _info "Using servercow"
  _debug fulldomain "$fulldomain"
  _debug txtvalue "$fulldomain"

  SERVERCOW_API_Username="${SERVERCOW_API_Username:-$(_readaccountconf_mutable SERVERCOW_API_Username)}"
  SERVERCOW_API_Password="${SERVERCOW_API_Password:-$(_readaccountconf_mutable SERVERCOW_API_Password)}"
  if [ -z "$SERVERCOW_API_Username" ] || [ -z "$SERVERCOW_API_Password" ]; then
    SERVERCOW_API_Username=""
    SERVERCOW_API_Password=""
    _err "You don't specify servercow api username and password yet."
    _err "Please create your username and password and try again."
    return 1
  fi

  _debug "First detect the root zone"
  if ! _get_root "$fulldomain"; then
    _err "invalid domain"
    return 1
  fi

  _debug _sub_domain "$_sub_domain"
  _debug _domain "$_domain"

  if _servercow_api DELETE "$_domain" "{\"type\":\"TXT\",\"name\":\"$fulldomain\"}"; then
    if printf -- "%s" "$response" | grep "ok" >/dev/null; then
      _info "Deleted, OK"
      _contains "$response" '"message":"ok"'
    else
      _err "delete txt record error."
      return 1
    fi
  fi

}

####################  Private functions below ##################################

# _acme-challenge.www.domain.com
# returns
#  _sub_domain=_acme-challenge.www
#  _domain=domain.com
_get_root() {
  fulldomain=$1
  i=2
  p=1

  while true; do
    _domain=$(printf "%s" "$fulldomain" | cut -d . -f "$i"-100)

    _debug _domain "$_domain"
    if [ -z "$_domain" ]; then
      # not valid
      return 1
    fi

    if ! _servercow_api GET "$_domain"; then
      return 1
    fi

    if ! _contains "$response" '"error":"no such domain in user context"' >/dev/null; then
      _sub_domain=$(printf "%s" "$fulldomain" | cut -d . -f 1-"$p")
      if [ -z "$_sub_domain" ]; then
        # not valid
        return 1
      fi

      return 0
    fi

    p=$i
    i=$(_math "$i" + 1)
  done

  return 1
}

_servercow_api() {
  method=$1
  domain=$2
  data="$3"

  export _H1="Content-Type: application/json"
  export _H2="X-Auth-Username: $SERVERCOW_API_Username"
  export _H3="X-Auth-Password: $SERVERCOW_API_Password"

  if [ "$method" != "GET" ]; then
    _debug data "$data"
    response="$(_post "$data" "$SERVERCOW_API/$domain" "" "$method")"
  else
    response="$(_get "$SERVERCOW_API/$domain")"
  fi

  if [ "$?" != "0" ]; then
    _err "error $domain"
    return 1
  fi
  _debug2 response "$response"
  return 0
}
