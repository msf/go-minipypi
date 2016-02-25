# go-minipypi

Bare minimums implementation of a Pypi server that proxies all requests to an S3 bucket.

This was implemented by looking what was required to get pip install commands such as:
> pip install -v  --no-index --find-links=http://localhost:8080/ -r requirements.txt
working.

It does the job.
