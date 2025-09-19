#!/bin/bash
set -e

# try to get a tag that exactly matches HEAD
TAG="$(git describe --exact-match --tags HEAD 2>/dev/null || true)"

if [ -n "$TAG" ]; then
    APP_VERSION="${TAG#v}"
else
    GIT_ROOT=$(git rev-parse --show-toplevel)
    APP_VERSION=$(cat "${GIT_ROOT}/VERSION")
fi

echo "VERSION: $VERSION"

# --- read only what we need from /etc/os-release without polluting ---
if [ -r /etc/os-release ]; then
    OS_ID=$(awk -F= '/^ID=/{gsub(/"/,"",$2);print tolower($2)}' /etc/os-release)
    OS_VERSION_ID=$(awk -F= '/^VERSION_ID=/{gsub(/"/,"",$2);print $2}' /etc/os-release)
else
    OS_ID="unknown"
    OS_VERSION_ID=""
fi

case "$OS_ID" in
    debian) DIST_SUFFIX="deb${OS_VERSION_ID}u1" ;;
    ubuntu) DIST_SUFFIX="ubuntu${OS_VERSION_ID}u1" ;;
    *)      DIST_SUFFIX="${OS_ID}${OS_VERSION_ID}u1" ;;
esac

COMMIT=$(git rev-parse --short HEAD)
git diff --quiet --exit-code || COMMIT="$COMMIT.dirty"
JENKINS_BUILD_NUMBER=${BUILD_NUMBER:-0}

echo -n "${APP_VERSION}+${JENKINS_BUILD_NUMBER}.${COMMIT}-${DIST_SUFFIX}" > generated_version

