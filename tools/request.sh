#!/bin/bash

request=$1
service=$2
count=$3

if ! [[ "$count" =~ ^[0-9]+$ ]]; then
   count=1
fi

for (( i=1; i<=$count; i++ )); do
  curl --header "Content-Type: application/json" \
    --request POST \
    --data "{\"url\": \"$request\"}" \
    $service
done
