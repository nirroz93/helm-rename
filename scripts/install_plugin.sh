#!/bin/sh -e

if [ -n "${HELM_LINTER_PLUGIN_NO_INSTALL_HOOK}" ]; then
    echo "Development mode: not downloading versioned release."
    exit 0
fi

# shellcheck disable=SC2002
version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"
echo "Downloading and installing helm-edit v${version} ..."

url=""
if [ "$(uname)" = "Darwin" ]; then
    if [ "$(uname -m)" = "arm64" ]; then
        url="https://github.com/nirroz93/helm-edit/releases/download/v${version}/helm-edit_${version}_darwin_arm64.tar.gz"
    else
        url="https://github.com/nirroz93/helm-edit/releases/download/v${version}/helm-edit_${version}_darwin_amd64.tar.gz"
    fi
elif [ "$(uname)" = "Linux" ] ; then
    if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
        url="https://github.com/nirroz93/helm-edit/releases/download/v${version}/helm-edit_${version}_linux_arm64.tar.gz"
    else
        url="https://github.com/nirroz93/helm-edit/releases/download/v${version}/helm-edit_${version}_linux_amd64.tar.gz"
    fi
fi

echo "$url"

mkdir -p "bin"
mkdir -p "releases/v${version}"

# Download with curl if possible.
# shellcheck disable=SC2230
if [ -x "$(which curl 2>/dev/null)" ]; then
    curl -sSL "${url}" -o "releases/v${version}.tar.gz"
else
    wget -q "${url}" -O "releases/v${version}.tar.gz"
fi
tar xzf "releases/v${version}.tar.gz" -C "releases/v${version}"
mv "releases/v${version}/rename" "bin/rename" || \
mv "releases/v${version}/completion.yaml" .
mv "releases/v${version}/plugin.yaml" .
mv "releases/v${version}/README.md" .
mv "releases/v${version}/LICENSE" .