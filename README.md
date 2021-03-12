# ticker – Event Stream Service

## Getting started

Ticker compiles to a single binary that runs both the ticker server, and a generic client that allows to manage, 
test 
and monitor the ticker server.

### Starting the server

Ticker currently uses PostgreSQL as persistent data storage. Configure the database using the `--database` flag or 
the `TICKER_DATABASE` environment variable. You can specify the value as one of the options documented 
[here](https://github.com/lib/pq/blob/master/url.go). 

Ticker listens on TCP port 6677 by default, which you can change using the `--listen` flag.


