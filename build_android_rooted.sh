#!/usr/bin/env sh
#
# Builds static tailscaled and tailscale binaries for running natively
# (outside the Android app) on a rooted Android device, with their default
# state directory and socket at $STATE_DIR and $SOCKET instead of under
# /var, which Android doesn't have, then installs them into $INSTALL_DIR.
#
# It cross-compiles, so it works on the device (e.g. in a Linux chroot) or
# on any other machine:
#
#	./build_android_rooted.sh
#	STATE_DIR=/some/dir SOCKET=/some/dir/tailscaled.sock ./build_android_rooted.sh
#	GOARCH=arm STRIP=1 INSTALL_DIR= ./build_android_rooted.sh
#
# The binaries are written to cmd/tailscaled/tailscaled and
# cmd/tailscale/tailscale. When run in a chroot on the device, where
# $INSTALL_DIR isn't visible, they're installed through /proc/1/root (the
# root of init's mount namespace). Set INSTALL_DIR= to skip installing.

set -eu

cd "$(dirname "$0")"

STATE_DIR="${STATE_DIR:-/data/local/var/lib/tailscale}"
SOCKET="${SOCKET:-/data/local/var/run/tailscale/tailscaled.sock}"
INSTALL_DIR="${INSTALL_DIR-/data/local/bin}"

export TS_USE_TOOLCHAIN=1
eval "$(./build_dist.sh shellvars)"

# build_dist.sh can't take extra -ldflags (a second -ldflags would replace
# its version stamps), so stamp the version here as well.
ldflags="-X tailscale.com/version.longStamp=$VERSION_LONG"
ldflags="$ldflags -X tailscale.com/version.shortStamp=$VERSION_SHORT"
ldflags="$ldflags -X tailscale.com/paths.stateDirOverride=$STATE_DIR"
ldflags="$ldflags -X tailscale.com/paths.socketPathOverride=$SOCKET"
if [ "${STRIP:-0}" = 1 ]; then
	ldflags="$ldflags -s -w"
fi

export CGO_ENABLED=0 GOOS=linux GOARCH="${GOARCH:-arm64}"
for cmd in tailscaled tailscale; do
	echo "# building $cmd for $GOOS/$GOARCH (state $STATE_DIR, socket $SOCKET)" >&2
	./tool/go build -trimpath -ldflags "$ldflags" -o "cmd/$cmd/$cmd" "./cmd/$cmd"
done

[ -n "$INSTALL_DIR" ] || exit 0
dest="$INSTALL_DIR"
if [ ! -d "$dest" ] && [ -d "/proc/1/root$INSTALL_DIR" ]; then
	dest="/proc/1/root$INSTALL_DIR"
fi
if [ ! -d "$dest" ]; then
	echo "# $INSTALL_DIR not found; not installing" >&2
	exit 0
fi
for cmd in tailscaled tailscale; do
	# Copy then rename, so replacing a running tailscaled works.
	cp "cmd/$cmd/$cmd" "$dest/$cmd.new"
	chmod 755 "$dest/$cmd.new"
	mv "$dest/$cmd.new" "$dest/$cmd"
	echo "# installed $INSTALL_DIR/$cmd" >&2
done
