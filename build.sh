#!/bin/bash

APP="qaz"

for GOOS in darwin linux; do
  for GOARCH in 386 amd64; do
    echo "INFO: Compiling: ${APP} in ${GOOS}-${GOARCH}"
    echo "INFO: "go build -o "${APP}_${GOOS}_${GOARCH}" "${APP}"
    GOOS=${GOOS} GOARCH=${GOARCH} go build -o "${APP}_${GOOS}_${GOARCH}" .
  done
done