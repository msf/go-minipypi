go-minipypi
===========

Bare minimums implementation of a Pypi server that proxies all requests to an S3 bucket.

This was implemented by looking what was required to get pip install commands such as this one working:
> pip install -v  --no-index --find-links=http://localhost:8080/ -r requirements.txt

it does the job.


Instalation
-----------

1. Install go:

	```
	brew install go
	```

2. Setup your GOPATH and pull the code:

	```
	go get github.com/citymapper/go-minipypi
	```

3. Build the code:

	```
	cd $GOPATH/src/github.com/citymapper/go-minipypi
	go build .
	```

4. Run it:

	```
	./go-minipypi
	```

Notes:
------
Its requires a config file, see config.yml.
It requires a ini file that holds the AWS credentials. See aws_credentials.ini


