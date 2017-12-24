# Sequence ID generator as a micro-service

The sequence ID is designed to generate sequential ID for no-sql applications.

gRPC protocol is used to provide high efficiency network communication.

## Requirement

    go get -u google.golang.org/grpc
    go get -u github.com/golang/protobuf/protoc-gen-go
    go get -u github.com/go-sql-driver/mysql

## Install

    go get github.com/c9s/sid/sidserver

## Setting up

To setup the sid generator, you need few things:

1. DSN for MySQL server 
2. The keys of the sequences.

You need to create a config file like this:

```json
{
    "backend": {
        "mysql": {
            "dsn": "root@unix(/opt/local/var/run/mysql57/mysqld.sock)/sid"
        }
    },
    "sequences": {
        "jobs": {},
        "orders": {}
    }
}
```

## Running

Simply run the following commands to run the server:

    sid-server -config config.json

## License

MIT License

## Author

Yo-An Lin <yoanlin93@gmail.com>



