#!/bin/sh
# Gera cada template do catálogo. "dry" só simula; "test" gera com todos os serviços e roda os testes do projeto.
# Uso: scripts/templates.sh dry|test ./romanctl
set -eu

mode=${1:?use: templates.sh dry|test ./romanctl}
romanctl=${2:-romanctl}

"$romanctl" catalog --json | jq -c '.templates[]' | while read -r tpl; do
  stack=$(echo "$tpl" | jq -r .stack)
  id=$(echo "$tpl" | jq -r .id)
  arch=$(echo "$tpl" | jq -r '.arch | join(",")')
  with=$(echo "$tpl" | jq -r '.services // [] | join(",")')
  name="t-$stack-$id"
  echo "::group::$stack/$id"

  if [ "$mode" = dry ]; then
    "$romanctl" new "$name" --stack "$stack" --arch "$arch" ${with:+--with "$with"} --dry-run --json >/dev/null
    "$romanctl" new "$name" --stack "$stack" --arch "$arch" --dry-run --json >/dev/null
    echo "::endgroup::"
    continue
  fi

  requires=$(echo "$tpl" | jq -r '.requires // [] | join(" ")')
  if [ -n "$requires" ]; then "$romanctl" tools install $requires; fi
  for service in $(echo "$with" | tr ',' ' '); do
    "$romanctl" start "$service"
  done
  created=$("$romanctl" new "$name" --stack "$stack" --arch "$arch" ${with:+--with "$with"} --sem-git --json)
  echo "$created" | jq -e '.setup != "failed"' >/dev/null
  (cd "$(echo "$created" | jq -r .path)" && sh -c "$(echo "$created" | jq -r .manifest.test)")
  "$romanctl" rm "$name" --yes
  echo "::endgroup::"
done
