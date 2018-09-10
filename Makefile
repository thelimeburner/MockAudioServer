all: clean build test

build: 
	go build -o server.out

clean:
	rm -rf server.out

test:
	./server.out
