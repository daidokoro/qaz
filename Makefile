# This is how we want to name the binary output
BINARY=qaz

all:
	go build -o ${BINARY}

test:
	go test

clean:
	rm -f qaz

get-deps:
	go get ./...
