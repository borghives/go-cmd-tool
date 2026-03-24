build:
	go build -o sitedb ./sitedb
	go build -o sitesecret ./sitesecret

install:
	go install github.com/borghives/sitestate/sitedb@latest
	go install github.com/borghives/sitestate/sitesecret@latest

clean:
	rm -f sitedb
	rm -f sitedb.exe
	rm -f sitesecret
	rm -f sitesecret.exe