go-minipypi
===========

A bare minimum implementation of a PyPI server that proxies all requests to an S3 bucket.

This was implemented by looking what was required to get pip install commands such as this one working:
> pip install -v  --no-index --find-links=http://localhost:8080/ -r requirements.txt

Installation
------------

1. Install [go 1.11](https://golang.org/dl/) or later.

2. Clone this repository to somewhere outside of your GOPATH.

3. Build the code:

	```sh
	go build .
	```

4. Run it:

	```
	./go-minipypi
	```

Notes:
------
It requires a config file, see `config.yml`. Drop the `credentialsfile` parameter to use the [default AWS credentials chain](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials).
