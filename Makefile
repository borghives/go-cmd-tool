build:
	go build -o sitedb ./sitedb
	go build -o sitesecret ./sitesecret

install:
	go install ./sitedb
	go install ./sitesecret

clean:
	rm -f sitedb
	rm -f sitedb.exe
	rm -f sitesecret
	rm -f sitesecret.exe