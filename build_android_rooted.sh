#!/usr/bin/env sh
#
# Builds static tailscaled and tailscale binaries for running natively
# (outside the Android app) on a rooted Android device, with their default
# state directory and socket under $STATE_DIR instead of /var, which
# Android doesn't have.
#
# It cross-compiles, so it works on the device (e.g. in a Linux chroot) or
# on any other machine:
#
#	./build_android_rooted.sh
#	STATE_DIR=/some/dir GOARCH=arm STRIP=1 ./build_android_rooted.sh
#
# The binaries are written to cmd/tailscaled/tailscaled and
# cmd/tailscale/tailscale.

set -eu

cd "$(dirname "$0")"

STATE_DIR="${STATE_DIR:-/data/local/lib/tailscale}"

export TS_USE_TOOLCHAIN=1
eval "$(./build_dist.sh shellvars)"

# build_dist.sh can't take extra -ldflags (a second -ldflags would replace
# its version stamps), so stamp the version here as well.
ldflags="-X tailscale.com/version.longStamp=$VERSION_LONG"
ldflags="$ldflags -X tailscale.com/version.shortStamp=$VERSION_SHORT"
ldflags="$ldflags -X tailscale.com/paths.stateDirOverride=$STATE_DIR"
if [ "${STRIP:-0}" = 1 ]; then
	ldflags="$ldflags -s -w"
fi

export CGO_ENABLED=0 GOOS=linux GOARCH="${GOARCH:-arm64}"
for cmd in tailscaled tailscale; do
	echo "# building $cmd for $GOOS/$GOARCH, state dir $STATE_DIR" >&2
	./tool/go build -trimpath -ldflags "$ldflags" -o "cmd/$cmd/$cmd" "./cmd/$cmd"
done
