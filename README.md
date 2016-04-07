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

2. Setup your GOPATH env variable and pull the code:

	```
	go get github.com/citymapper/go-minipypi
	```

3. Install the requirements (if you haven't done so already)

	```
	go get github.com/aws/aws-sdk-go/aws
	go get gopkg.in/yaml.v2
	```

4. Build the code:

	```
	cd $GOPATH/src/github.com/citymapper/go-minipypi
	go build .
	```

5. Run it:

	```
	./go-minipypi
	```

Notes:
------
Its requires a config file, see config.yml.
It requires a ini file that holds the AWS credentials. See aws_credentials.ini
