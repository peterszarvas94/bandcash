#!/usr/bin/env sh

set -eu

if [ "$#" -lt 3 ]; then
  echo "usage: $0 <env-file> -- <command...>" >&2
  exit 1
fi

env_file="$1"
shift

if [ "$1" != "--" ]; then
  echo "expected '--' before command" >&2
  exit 1
fi
shift

if [ ! -f "$env_file" ]; then
  echo "env file not found: $env_file" >&2
  exit 1
fi

source_file="$env_file"
case "$source_file" in
  */*) ;;
  *) source_file="./$source_file" ;;
esac

set -a
. "$source_file"
set +a

if [ -z "${APP_ENV:-}" ]; then
  case "$env_file" in
    *.production.local) APP_ENV="production" ;;
    *) APP_ENV="development" ;;
  esac
  export APP_ENV
fi

exec varlock run -- "$@"
