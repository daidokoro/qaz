#!/bin/bash

#
# We upload files only for tagged commits
#

if [ "$TRAVIS_TAG" = "" ]; then
  echo "There is no TRAVIS_TAG"
  exit 0
fi

if [ "$GITHUB_TOKEN" = "" ]; then
  echo "There is no GITHUB_TOKEN"
  exit 0
fi

echo "Go get ghr"
go get -u github.com/tcnksm/ghr

echo "Pushing changes to github."

upload_file() {
  FILENAME=$1
  for _ in {1..10}; do
    ghr --username daidokoro --repository qaz -t $GITHUB_TOKEN $TRAVIS_TAG $FILENAME
    if [ $? -eq 0 ]; then
      return
    fi
    echo "Upload failed, repeat"
    sleep 5
  done
}

upload_file "bin/shipa_linux_386"
upload_file "bin/shipa_linux_amd64"
upload_file "bin/shipa_darwin_386"
upload_file "bin/shipa_darwin_amd64"
