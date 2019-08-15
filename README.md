# Go Redis Implementation

The project is aiming to implement redis in golang

## Implemented functionality

- Transaction

  Supported transaction commands: watch, unwatch, multi, discard, exec

- Pipeline

  Really efficient according to benchmark result

- Timeout

  Set value with argument EX/PX

  Priority queue to pop the expired key-value

- Multiple database

- Multiple data type

  Supported data types: string, binary, integer

  implementation of more data types like linked list, hash dict, set etc. is on the way just after fundamental functionality implementation.

- Client btw

  Client can support all commands with server library

- Basic read/write

  Supported read/write commands: set, get, incr, desc etc.

- Single-threaded server

- Resonable TCP protocol

  Protocol is similar to the one of original redis


# TODO

- Sub / Pub

- AOF

- RDB

- More detailed commands e.g. ttl, expire etc.

- More data types e.g. set, ordered set, hash etc.
