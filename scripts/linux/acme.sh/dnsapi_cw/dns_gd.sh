#!/usr/bin/env sh

ABS_CURR_PATH=$(dirname $(realpath "${BASH_SOURCE[0]}"))
SRC_FILE="${ABS_CURR_PATH}/../acme_src.sh"
. "${SRC_FILE}"

# shellcheck disable=SC2034
dns_gd_info='GoDaddy.com
Site: GoDaddy.com
Docs: github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_gd
Options:
 GD_Key API Key
 GD_Secret API Secret
'

GD_Api="https://api.godaddy.com/v1"

########  Public functions #####################

#Usage: add  _acme-challenge.www.domain.com   "XKrxpRBosdIKFzxW_CT3KLZNf6q0HG9i01zxXp5CPBs"
dns_gd_add() {
  fulldomain=$1
  txtvalue=$2

  GD_Key="${GD_Key:-$(_readaccountconf_mutable GD_Key)}"
  GD_Secret="${GD_Secret:-$(_readaccountconf_mutable GD_Secret)}"
  if [ -z "$GD_Key" ] || [ -z "$GD_Secret" ]; then
    GD_Key=""
    GD_Secret=""
    _err "You didn't specify godaddy api key and secret yet."
    _err "Please create your key and try again."
    return 1
  fi

  #save the api key and email to the account conf file.
  _saveaccountconf_mutable GD_Key "$GD_Key"
  _saveaccountconf_mutable GD_Secret "$GD_Secret"

  _debug "First detect the root zone"
  if ! _get_root "$fulldomain"; then
    _err "invalid domain"
    return 1
  fi

  _debug _sub_domain "$_sub_domain"
  _debug _domain "$_domain"

  _debug "Getting existing records"
  if ! _gd_rest GET "domains/$_domain/records/TXT/$_sub_domain"; then
    return 1
  fi

  if _contains "$response" "$txtvalue"; then
    _info "This record already exists, skipping"
    return 0
  fi

  _add_data="{\"data\":\"$txtvalue\"}"
  for t in $(echo "$response" | tr '{' "\n" | grep "\"name\":\"$_sub_domain\"" | tr ',' "\n" | grep '"data"' | cut -d : -f 2); do
    _debug2 t "$t"
    # ignore empty (previously removed) records, to prevent useless _acme-challenge TXT entries
    if [ "$t" ] && [ "$t" != '""' ]; then
      _add_data="$_add_data,{\"data\":$t}"
    fi
  done
  _debug2 _add_data "$_add_data"

  _info "Adding record"
  if _gd_rest PUT "domains/$_domain/records/TXT/$_sub_domain" "[$_add_data]"; then
    _debug "Checking updated records of '${fulldomain}'"

    if ! _gd_rest GET "domains/$_domain/records/TXT/$_sub_domain"; then
      _err "Validating TXT record for '${fulldomain}' with rest error [$?]." "$response"
      return 1
    fi

    if ! _contains "$response" "$txtvalue"; then
      _err "TXT record '${txtvalue}' for '${fulldomain}', value wasn't set!"
      return 1
    fi
  else
    _err "Add txt record error, value '${txtvalue}' for '${fulldomain}' was not set."
    return 1
  fi

  _sleep 10
  _info "Added TXT record '${txtvalue}' for '${fulldomain}'."
  return 0
}

#fulldomain
dns_gd_rm() {
  fulldomain=$1
  txtvalue=$2

  GD_Key="${GD_Key:-$(_readaccountconf_mutable GD_Key)}"
  GD_Secret="${GD_Secret:-$(_readaccountconf_mutable GD_Secret)}"

  _debug "First detect the root zone"
  if ! _get_root "$fulldomain"; then
    _err "invalid domain"
    return 1
  fi

  _debug _sub_domain "$_sub_domain"
  _debug _domain "$_domain"

  _debug "Getting existing records"
  if ! _gd_rest GET "domains/$_domain/records/TXT/$_sub_domain"; then
    return 1
  fi

  if ! _contains "$response" "$txtvalue"; then
    _info "The record does not exist, skip"
    return 0
  fi

  _add_data=""
  for t in $(echo "$response" | tr '{' "\n" | grep "\"name\":\"$_sub_domain\"" | tr ',' "\n" | grep '"data"' | cut -d : -f 2); do
    _debug2 t "$t"
    if [ "$t" ] && [ "$t" != "\"$txtvalue\"" ]; then
      if [ "$_add_data" ]; then
        _add_data="$_add_data,{\"data\":$t}"
      else
        _add_data="{\"data\":$t}"
      fi
    fi
  done
  if [ -z "$_add_data" ]; then
    # delete empty record
    _debug "Delete last record for '${fulldomain}'"
    if ! _gd_rest DELETE "domains/$_domain/records/TXT/$_sub_domain"; then
      _err "Cannot delete empty TXT record for '$fulldomain'"
      return 1
    fi
  else
    # remove specific TXT value, keeping other entries
    _debug2 _add_data "$_add_data"
    if ! _gd_rest PUT "domains/$_domain/records/TXT/$_sub_domain" "[$_add_data]"; then
      _err "Cannot update TXT record for '$fulldomain'"
      return 1
    fi
  fi
}

####################  Private functions below ##################################
#_acme-challenge.www.domain.com
#returns
# _sub_domain=_acme-challenge.www
# _domain=domain.com
_get_root() {
  domain=$1
  i=2
  p=1
  while true; do
    h=$(printf "%s" "$domain" | cut -d . -f "$i"-100)
    if [ -z "$h" ]; then
      #not valid
      return 1
    fi

    if ! _gd_rest GET "domains/$h"; then
      return 1
    fi

    if _contains "$response" '"code":"NOT_FOUND"'; then
      _debug "$h not found"
    else
      _sub_domain=$(printf "%s" "$domain" | cut -d . -f 1-"$p")
      _domain="$h"
      return 0
    fi
    p="$i"
    i=$(_math "$i" + 1)
  done
  return 1
}

_gd_rest() {
  m=$1
  ep="$2"
  data="$3"
  _debug "$ep"

  export _H1="Authorization: sso-key $GD_Key:$GD_Secret"
  export _H2="Content-Type: application/json"

  if [ "$data" ] || [ "$m" = "DELETE" ]; then
    _debug "data ($m): " "$data"
    response="$(_post "$data" "$GD_Api/$ep" "" "$m")"
  else
    response="$(_get "$GD_Api/$ep")"
  fi

  if [ "$?" != "0" ]; then
    _err "error on rest call ($m): $ep"
    return 1
  fi
  _debug2 response "$response"
  if _contains "$response" "UNABLE_TO_AUTHENTICATE"; then
    _err "It seems that your api key or secret is not correct."
    return 1
  fi
  return 0
}
