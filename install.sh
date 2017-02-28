#!/bin/sh

install () {
  set -eu

  UNAME=$(uname)
  if [ "$UNAME" != "Linux" -a "$UNAME" != "Darwin" -a "$UNAME" != "OpenBSD" ] ; then
      echo "Sorry, OS not supported: ${UNAME}. Download binary from https://github.com/daidokoro/qaz/releases"
      exit 1
  fi

  case "${UNAME}" in
    Darwin)
      OSX_ARCH=$(uname -m)
      if [ "${OSX_ARCH}" = "x86_64" ] ; then
        PLATFORM="darwin_amd64"
      else
        echo "Sorry, architecture not supported: ${OSX_ARCH}. Download binary from https://github.com/daidokoro/qaz/releases"
        exit 1
      fi

    ;;

    Linux)
      LINUX_ARCH=$(uname -m)
      if [ "${LINUX_ARCH}" = "i686" ] ; then
        PLATFORM="linux_386"
      elif [ "${LINUX_ARCH}" = "x86_64" ] ; then
        PLATFORM="linux_amd64"
      else
        echo "Sorry, architecture not supported: ${LINUX_ARCH}. Download binary from https://github.com/daidokoro/qaz/releases"
        exit 1
      fi
    ;;
  esac

  LATEST=$(curl -s https://api.github.com/repos/daidokoro/qaz/tags | grep -Eo '"name":.*[^\\]",'  | head -n 1 | sed 's/[," ]//g' | cut -d ':' -f 2)
  URL="https://github.com/daidokoro/qaz/releases/download/$LATEST/qaz_$PLATFORM"
  DEST=${DEST:-/usr/local/bin/qaz}

  if [ -z $LATEST ] ; then
    echo "Error requesting. Download binary from https://github.com/daidokoro/qaz/releases"
    exit 1
  else
    echo "Downloading Qaz binary from https://github.com/daidokoro/qaz/releases/download/$LATEST/qaz_$PLATFORM to $DEST"
    if curl -sL https://github.com/daidokoro/qaz/releases/download/$LATEST/qaz_$PLATFORM -o $DEST; then
      chmod +x $DEST
      echo "Qaz installation was successful"
    else
      echo "Installation failed. You may need elevated permissions."
    fi
  fi
}

install
