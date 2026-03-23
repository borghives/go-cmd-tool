build:
	go build -o sitedb ./sitedb

install:
	go install github.com/borghives/sitestate/sitedb@latest

clean:
	rm -f sitedb
	rm -f sitedb.exe