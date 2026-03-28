#!/bin/sh
set -e

on_exit () {
	[ $? -eq 0 ] && exit
	echo 'ERROR: Feature "Node.js (via nvm), yarn and pnpm." (ghcr.io/devcontainers/features/node) failed to install! Look at the documentation at ${documentation} for help troubleshooting this error.'
}

trap on_exit EXIT

set -a
. ../devcontainer-features.builtin.env
. ./devcontainer-features.env
set +a

echo ===========================================================================

echo 'Feature       : Node.js (via nvm), yarn and pnpm.'
echo 'Description   : Installs Node.js, nvm, yarn, pnpm, and needed dependencies.'
echo 'Id            : ghcr.io/devcontainers/features/node'
echo 'Version       : 1.7.1'
echo 'Documentation : https://github.com/devcontainers/features/tree/main/src/node'
echo 'Options       :'
echo '    INSTALLYARNUSINGAPT="false"
    NODEGYPDEPENDENCIES="true"
    NVMINSTALLPATH="/usr/local/share/nvm"
    NVMVERSION="latest"
    PNPMVERSION="latest"
    VERSION="lts"'
echo 'Environment   :'
printenv
echo ===========================================================================

chmod +x ./install.sh
./install.sh
