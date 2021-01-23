# Definition of the Ticker Protocol

Ticker uses a text-based protocol

Using arrows to represent the direction of the message (`>` client to server; `<` server to client)

## Initializing a connection
```
< INFO 1
{"server_version": "0.1.0"}
```

## Example Producer Connections

### Append Events 
```
> APPEND de.customer.12345 create-customer 18 d3b09abe30cfe2edff4ee9e0a141c93bf5b3af87 
> {"hello": "world"}
< APPENDED 123
```

### Store Snapshots
```
> STORE de.customer.12345 123 18 d3b09abe30cfe2edff4ee9e0a141c93bf5b3af87
> {"hello": "world"}
< STORED 123
```

## Example Consumer Connections

### Subscribe to all Events in realtime 
```
> SUBSCRIBE de.customer.*
< STARTING 123
< EVENT 123 de.customer.12345 create-customer 18 d3b09abe30cfe2edff4ee9e0a141c93bf5b3af87
< {"hello": "world"}
```

### Statefully stream all Events 
```
> STREAM client1 de.customer.*
< STARTING 123
< SNAPSHOT 123 de.customer.12345 18 d3b09abe30cfe2edff4ee9e0a141c93bf5b3af87
< {"hello": "world"}
> NEXT
< EVENT 124 de.customer.12346 create-customer 18 d3b09abe30cfe2edff4ee9e0a141c93bf5b3af87
< {"hello": "world"}
> ACK 123
> ACK 124
> STOP
< STOPPED
```

## Messages

`APPEND <aggregate> <type> <length> <sha1sum>`

`APPENDED <sequence>`

`STORE <aggregate> <type> <sequence> <length> <sha1sum>`
(Not yet implemented)

`STORED <sequence>`
(Not yet implemented)

`BURY <aggregate>`
(Not yet implemented)

`BURIED <aggregate>`
(Not yet implemented)

`SUBSCRIBE <aggregate-pattern> [types...]`

`STREAM <client-id> <aggregate-pattern> [types...]`

`STARTING <client-id> <sequence>`

`EVENT <sequence> <aggregate> <type> <length> <sha1sum>`

`SNAPSHOT <sequence> <aggregate> <length> <sha1sum>`
(Not yet implemented)

`TOMBSTONE <sequence> <aggregate>`
(Not yet implemented)

`NEXT`
(Not yet implemented)

`ACK [sequence]`

`REPEAT [sequence]`
(Not yet implemented)

`QUIT`
