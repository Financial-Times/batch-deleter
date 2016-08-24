# Batch Deleter

__An API for doing batch deletes across all nodes. For example, to delete a list of memberships from
all the neo4j instances in prod__

# Usage
Run locally on any port.

`$GOPATH\bin\batch-deleter --port=8080`

Make a POST request to the `\batchdelete` endpoint, providing a body that contains:
* hosts: a list, each value being the scheme and route to the service, e.g. `https://pre-prod-up.ft.com/__memberships-rw-neo4j-red`
* path: the resource path, e.g. `people`
* uuids: a list of uuids to delete

See the sample json files in this folder for more detailed examples.

If running this against a set of hosts that require authorization, specify the basic auth username and password
on the request and it will be passed on.

Batch deleter will work for any app that supports a DELETE request and returns either 204 (for successful delete)
or 404 (if nothing found for delete).

It will log out any responses that fail for any of these reasons:
* the body is gzipped but can't be extracted
* the body isn't valid json or can't be decoded
* the request URL can't be created
* there's an error when attempting to execute the http request
* the status code isn't valid. For example, if a 401 (not authorized) is returned

NB: this is a utility and there are no tests! 

## Installation

* `go get github.com/Financial-Times/batch-deleter`
* `cd $GOPATH/src/github.com/Financial-Times/batch-deleter`
* `go install`
