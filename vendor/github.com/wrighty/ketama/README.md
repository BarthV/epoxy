[![Build Status](https://travis-ci.org/wrighty/ketama.svg?branch=master)](https://travis-ci.org/wrighty/ketama)
[![Code Climate](https://codeclimate.com/github/wrighty/ketama/badges/gpa.svg)](https://codeclimate.com/github/wrighty/ketama)

# ketama

A port of [libketama](https://github.com/RJ/ketama/), written in Go. Provides
consistent hashing of keys across a cluster of nodes (for example, memcached
servers) and is resilient in the event of a node being removed. Nodes are
projected across a 32 bit continuum (usually 160 times per node). When hashing a
key to select a server, calculate the point where the key is on the continuum
and then find the nearest node that is just greater than the key's point. When a
node is removed only those sections of the continuum that hashed to that node
are redistributed across the remaining nodes, rather than rehashing the entire
keyspace amongst the remaining nodes.

Features extensive testing, including exact compatibility with the original
[C implementation](https://github.com/RJ/ketama/tree/master/libketama), see
`TestCompatibilityWithLibketama` for details. This means it can coexist in a
mixed language environment where node selection must be consistent between
applications. MD5 is used throughout for hashing.

# Usage

To install:

```
go get github.com/wrighty/ketama
```

To use, all hosts are equally weighted:

```go
hosts := []string{"host1", "host2"}
c := ketama.Make(hosts)
host := c.GetHost('my key')
```

Or with weighted hosts:

```go
hosts := map[string]uint{
    "host1": 79,
    "host2": 1,
}
c := ketama.MakeWithWeights(hosts)
host := c.GetHost('my key')
```

