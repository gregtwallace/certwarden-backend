#!/usr/bin/env sh

ABS_CURR_PATH=$(dirname $(realpath "${BASH_SOURCE[0]}"))
SRC_FILE="${ABS_CURR_PATH}/../acme_src.sh"
. "${SRC_FILE}"

# shellcheck disable=SC2034
dns_rackcorp_info='RackCorp.com
Site: RackCorp.com
Docs: github.com/acmesh-official/acme.sh/wiki/dnsapi2#dns_rackcorp
Options:
 RACKCORP_APIUUID API UUID. See Portal: ADMINISTRATION -> API
 RACKCORP_APISECRET API Secret
Issues: github.com/acmesh-official/acme.sh/issues/3351
Author: Stephen Dendtler <sdendtler@rackcorp.com>
'

RACKCORP_API_ENDPOINT="https://api.rackcorp.net/api/rest/v2.4/json.php"

########  Public functions #####################

dns_rackcorp_add() {
  fulldomain="$1"
  txtvalue="$2"

  _debug fulldomain="$fulldomain"
  _debug txtvalue="$txtvalue"

  if ! _rackcorp_validate; then
    return 1
  fi

  _debug "Searching for root zone"
  if ! _get_root "$fulldomain"; then
    return 1
  fi
  _debug _lookup "$_lookup"
  _debug _domain "$_domain"

  _info "Creating TXT record."

  if ! _rackcorp_api dns.record.create "\"name\":\"$_domain\",\"type\":\"TXT\",\"lookup\":\"$_lookup\",\"data\":\"$txtvalue\",\"ttl\":300"; then
    return 1
  fi

  return 0
}

#Usage: fulldomain txtvalue
#Remove the txt record after validation.
dns_rackcorp_rm() {
  fulldomain=$1
  txtvalue=$2

  _debug fulldomain="$fulldomain"
  _debug txtvalue="$txtvalue"

  if ! _rackcorp_validate; then
    return 1
  fi

  _debug "Searching for root zone"
  if ! _get_root "$fulldomain"; then
    return 1
  fi
  _debug _lookup "$_lookup"
  _debug _domain "$_domain"

  _info "Creating TXT record."

  if ! _rackcorp_api dns.record.delete "\"name\":\"$_domain\",\"type\":\"TXT\",\"lookup\":\"$_lookup\",\"data\":\"$txtvalue\""; then
    return 1
  fi

  return 0
}

####################  Private functions below ##################################
#_acme-challenge.domain.com
#returns
# _lookup=_acme-challenge
# _domain=domain.com
_get_root() {
  domain=$1
  i=1
  p=1
  if ! _rackcorp_api dns.domain.getall "\"name\":\"$domain\""; then
    return 1
  fi
  while true; do
    h=$(printf "%s" "$domain" | cut -d . -f "$i"-100)
    _debug searchhost "$h"
    if [ -z "$h" ]; then
      _err "Could not find domain for record $domain in RackCorp using the provided credentials"
      #not valid
      return 1
    fi

    _rackcorp_api dns.domain.getall "\"exactName\":\"$h\""

    if _contains "$response" "\"matches\":1"; then
      if _contains "$response" "\"name\":\"$h\""; then
        _lookup=$(printf "%s" "$domain" | cut -d . -f 1-"$p")
        _domain="$h"
        return 0
      fi
    fi
    p=$i
    i=$(_math "$i" + 1)
  done

  return 1
}

_rackcorp_validate() {
  RACKCORP_APIUUID="${RACKCORP_APIUUID:-$(_readaccountconf_mutable RACKCORP_APIUUID)}"
  if [ -z "$RACKCORP_APIUUID" ]; then
    RACKCORP_APIUUID=""
    _err "You require a RackCorp API UUID (export RACKCORP_APIUUID=\"<api uuid>\")"
    _err "Please login to the portal and create an API key and try again."
    return 1
  fi

  _saveaccountconf_mutable RACKCORP_APIUUID "$RACKCORP_APIUUID"

  RACKCORP_APISECRET="${RACKCORP_APISECRET:-$(_readaccountconf_mutable RACKCORP_APISECRET)}"
  if [ -z "$RACKCORP_APISECRET" ]; then
    RACKCORP_APISECRET=""
    _err "You require a RackCorp API secret (export RACKCORP_APISECRET=\"<api secret>\")"
    _err "Please login to the portal and create an API key and try again."
    return 1
  fi

  _saveaccountconf_mutable RACKCORP_APISECRET "$RACKCORP_APISECRET"

  return 0
}
_rackcorp_api() {
  _rackcorpcmd=$1
  _rackcorpinputdata=$2
  _debug cmd "$_rackcorpcmd $_rackcorpinputdata"

  export _H1="Accept: application/json"
  response="$(_post "{\"APIUUID\":\"$RACKCORP_APIUUID\",\"APISECRET\":\"$RACKCORP_APISECRET\",\"cmd\":\"$_rackcorpcmd\",$_rackcorpinputdata}" "$RACKCORP_API_ENDPOINT" "" "POST")"

  if [ "$?" != "0" ]; then
    _err "error $response"
    return 1
  fi
  _debug2 response "$response"
  if _contains "$response" "\"code\":\"OK\""; then
    _debug code "OK"
  else
    _debug code "FAILED"
    response=""
    return 1
  fi
  return 0
}
