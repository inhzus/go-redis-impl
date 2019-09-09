# Go Redis Implementation

The project is aiming to implement redis in golang

## Implemented functionality

- Persistence
  
  The server automatically records every single value-changing command to AOF file and frequently clones the whole data to file

- Transaction

  Supported transaction commands: watch, unwatch, multi, discard, exec

- Pipeline

  Efficient according to benchmark results

- Timeout

  Set value with argument EX/PX

  Priority queue to pop the expired key-value

- Multiple database

- Multiple data type

  Supported data types: string, binary, integer

  Implementation of more data types like linked list, hash dict, set etc. is on the way just after fundamental functionality implementation

- Client btw

  The client can support all commands with server library

- CLI

- Basic read/write

  Supported read/write commands: set, get, incr, desc etc.

- Single-threaded server

- Reasonable TCP protocol

  Protocol is similar to the one of original redis

# TODO

1. Sub / Pub

2. Testcase for persistence

3. More detailed commands e.g. ttl, expire etc.

4. More data types e.g. set, ordered set, hash etc.
