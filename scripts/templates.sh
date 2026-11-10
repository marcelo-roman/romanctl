#!/bin/sh
# Percorre os templates do catálogo.
#   dry      simula a geração, com e sem serviços
#   inplace  roda o preparo e os testes de cada template na própria pasta, com os nomes de exemplo
#   test     gera com todos os serviços, cria o banco e roda os testes do projeto gerado
# Uso: scripts/templates.sh dry|inplace|test ./romanctl [stack/id]
set -eu

mode=${1:?use: templates.sh dry|inplace|test ./romanctl}
romanctl=${2:-romanctl}
only=${3:-}

"$romanctl" catalog --json | jq -c '.templates[]' | while read -r tpl; do
  stack=$(echo "$tpl" | jq -r .stack)
  id=$(echo "$tpl" | jq -r .id)
  arch=$(echo "$tpl" | jq -r '.arch | join(",")')
  with=$(echo "$tpl" | jq -r '.services // [] | join(",")')
  name="t-$stack-$id"
  if [ -n "$only" ] && [ "$only" != "$stack/$id" ]; then continue; fi
  echo "::group::$stack/$id"

  if [ "$mode" = dry ]; then
    "$romanctl" new "$name" --stack "$stack" --arch "$arch" ${with:+--with "$with"} --dry-run --json >/dev/null
    "$romanctl" new "$name" --stack "$stack" --arch "$arch" --dry-run --json >/dev/null
    echo "::endgroup::"
    continue
  fi

  requires=$(echo "$tpl" | jq -r '.requires // [] | join(" ")')
  if [ "$mode" = inplace ]; then
    if [ -n "$requires" ]; then "$romanctl" tools install $requires; fi
    root=$("$romanctl" templates status --json | jq -r .path)
    (
      cd "$root/$stack/projects/$id"
      echo "$tpl" | jq -r '.setup // [] | .[] | select((.when // "") | startswith("!") | not) | .run' | while read -r step; do sh -c "$step"; done
      sh -c "$(echo "$tpl" | jq -r .test)"
    )
    echo "::endgroup::"
    continue
  fi
  if [ -n "$requires" ]; then "$romanctl" tools install $requires; fi
  for service in $(echo "$with" | tr ',' ' '); do
    "$romanctl" install "$service"
    "$romanctl" start "$service"
  done
  created=$("$romanctl" new "$name" --stack "$stack" --arch "$arch" ${with:+--with "$with"} --sem-git --json)
  echo "$created" | jq -e '.setup != "failed"' >/dev/null
  "$romanctl" up "$name" --no-run --json >/dev/null
  dir=$(echo "$created" | jq -r .path)
  (cd "$dir" && set -a && { [ ! -f .env ] || . ./.env; } && set +a && sh -c "$(echo "$created" | jq -r .manifest.test)")
  "$romanctl" down "$name" --keep-services --json >/dev/null
  "$romanctl" rm "$name" --yes
  echo "::endgroup::"
done
