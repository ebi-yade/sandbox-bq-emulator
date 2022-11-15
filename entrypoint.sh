#!/bin/bash

set -eux

exec reflex -r '\.go$' -s go run "$1"
