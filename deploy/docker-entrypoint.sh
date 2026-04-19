#!/bin/sh
set -e
mkdir -p /data/repo /data/downloads 2>/dev/null || true

# GitHub SSH: host keys in /tmp so ~/.ssh can be mounted read-only (keys only).
if command -v ssh-keyscan >/dev/null 2>&1; then
	tmpkh=/tmp/warehouse_git_known_hosts
	ssh-keyscan -t rsa,ecdsa,ed25519 github.com >"$tmpkh" 2>/dev/null || true
	if [ -s "$tmpkh" ] && [ -z "$GIT_SSH_COMMAND" ]; then
		export GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=$tmpkh -o StrictHostKeyChecking=yes"
	fi
fi

# working_dir is often /data/repo (git mount); ./warehouse would resolve there and miss the image binary.
if [ "$1" = "./warehouse" ]; then
	shift
	exec /usr/local/bin/warehouse "$@"
fi

exec "$@"
