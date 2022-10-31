#!/usr/bin/env bash

set -o errexit
set -o nounset

# Solving all of this in bash was my initial idea and https://github.com/RomanNess/fundgrube-crawler/issues/1 inspired me to hack this initial idea.
SEARCH_REGEX=${SEARCH_REGEX:-".*"}

# create DB file for every search query so only new results are found
DB_FILE=/tmp/fundgrube-$(echo "${SEARCH_REGEX}" | md5sum | awk '{ print $1 }').tsv
touch "${DB_FILE}"

OUTLET_IDS=$(https 'https://www.saturn.de/de/data/fundgrube/api/postings?categorieIds=CAT_DE_SAT_786&limit=1&offset=0' | jq '.outlets[].id')

for outletId in ${OUTLET_IDS}; do
  https "https://www.saturn.de/de/data/fundgrube/api/postings?categorieIds=CAT_DE_SAT_786&limit=100&offset=0&outletIds=${outletId}" | jq -r '.postings[] | [.posting_id, .price, .name, .outlet.name] | @tsv' |
  while read -r postingId price name outletName; do
      # only print posting if it is not yet included in DB file
      grep -r "${postingId}" "${DB_FILE}" >/dev/null || echo -e "${postingId}\t${price}\t${name}\t${outletName}" | tee -a "${DB_FILE}"
  done
done
