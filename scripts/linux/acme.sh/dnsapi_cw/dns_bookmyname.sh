#!/usr/bin/env sh

ABS_CURR_PATH=$(dirname $(realpath "${BASH_SOURCE[0]}"))
SRC_FILE="${ABS_CURR_PATH}/../acme_src.sh"
. "${SRC_FILE}"

# shellcheck disable=SC2034
dns_bookmyname_info='BookMyName.com
Site: BookMyName.com
Docs: github.com/acmesh-official/acme.sh/wiki/dnsapi2#dns_bookmyname
Options:
 BOOKMYNAME_USERNAME Username
 BOOKMYNAME_PASSWORD Password
Issues: github.com/acmesh-official/acme.sh/issues/3209
Author: Neilpang
'

########  Public functions #####################

# BookMyName urls:
# https://BOOKMYNAME_USERNAME:BOOKMYNAME_PASSWORD@www.bookmyname.com/dyndns/?hostname=_acme-challenge.domain.tld&type=txt&ttl=300&do=add&value="XXXXXXXX"'
# https://BOOKMYNAME_USERNAME:BOOKMYNAME_PASSWORD@www.bookmyname.com/dyndns/?hostname=_acme-challenge.domain.tld&type=txt&ttl=300&do=remove&value="XXXXXXXX"'

# Output:
#good: update done, cid 123456, domain id 456789, type txt, ip XXXXXXXX
#good: remove done 1, cid 123456, domain id 456789, ttl 300, type txt, ip XXXXXXXX

# Be careful, BMN DNS servers can be slow to pick up changes; using dnssleep is thus advised.

# Usage:
# export BOOKMYNAME_USERNAME="ABCDE-FREE"
# export BOOKMYNAME_PASSWORD="MyPassword"
# /usr/local/ssl/acme.sh/acme.sh --dns dns_bookmyname --dnssleep 600 --issue -d domain.tld

#Usage: dns_bookmyname_add   _acme-challenge.www.domain.com   "XKrxpRBosdIKFzxW_CT3KLZNf6q0HG9i01zxXp5CPBs"
dns_bookmyname_add() {
  fulldomain=$1
  txtvalue=$2
  _info "Using bookmyname"
  _debug fulldomain "$fulldomain"
  _debug txtvalue "$txtvalue"

  BOOKMYNAME_USERNAME="${BOOKMYNAME_USERNAME:-$(_readaccountconf_mutable BOOKMYNAME_USERNAME)}"
  BOOKMYNAME_PASSWORD="${BOOKMYNAME_PASSWORD:-$(_readaccountconf_mutable BOOKMYNAME_PASSWORD)}"

  if [ -z "$BOOKMYNAME_USERNAME" ] || [ -z "$BOOKMYNAME_PASSWORD" ]; then
    BOOKMYNAME_USERNAME=""
    BOOKMYNAME_PASSWORD=""
    _err "You didn't specify BookMyName username and password yet."
    _err "Please specify them and try again."
    return 1
  fi

  #save the credentials to the account conf file.
  _saveaccountconf_mutable BOOKMYNAME_USERNAME "$BOOKMYNAME_USERNAME"
  _saveaccountconf_mutable BOOKMYNAME_PASSWORD "$BOOKMYNAME_PASSWORD"

  uri="https://${BOOKMYNAME_USERNAME}:${BOOKMYNAME_PASSWORD}@www.bookmyname.com/dyndns/"
  data="?hostname=${fulldomain}&type=TXT&ttl=300&do=add&value=${txtvalue}"
  result="$(_get "${uri}${data}")"
  _debug "Result: $result"

  if ! _startswith "$result" 'good: update done, cid '; then
    _err "Can't add $fulldomain"
    return 1
  fi

}

#Usage: fulldomain txtvalue
#Remove the txt record after validation.
dns_bookmyname_rm() {
  fulldomain=$1
  txtvalue=$2
  _info "Using bookmyname"
  _debug fulldomain "$fulldomain"
  _debug txtvalue "$txtvalue"

  BOOKMYNAME_USERNAME="${BOOKMYNAME_USERNAME:-$(_readaccountconf_mutable BOOKMYNAME_USERNAME)}"
  BOOKMYNAME_PASSWORD="${BOOKMYNAME_PASSWORD:-$(_readaccountconf_mutable BOOKMYNAME_PASSWORD)}"

  uri="https://${BOOKMYNAME_USERNAME}:${BOOKMYNAME_PASSWORD}@www.bookmyname.com/dyndns/"
  data="?hostname=${fulldomain}&type=TXT&ttl=300&do=remove&value=${txtvalue}"
  result="$(_get "${uri}${data}")"
  _debug "Result: $result"

  if ! _startswith "$result" 'good: remove done 1, cid '; then
    _info "Can't remove $fulldomain"
  fi

}

####################  Private functions below ##################################
