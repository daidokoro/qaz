# This is how we want to name the binary output
BINARY=qaze

all:
	go build -o ${BINARY}

test:
	go test

clean:
	rm -f qaze

get-deps:
	go get ./...
