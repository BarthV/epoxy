/*Package ketama is a port of libketama written in Go.

Provides consistent hashing of keys across a cluster of nodes (for example, memcached servers) and is resilient in the event of a node being removed.

Nodes are projected across a 32 bit continuum (usually 160 times per node). When hashing a key to select a server, calculate the point where the key is on the continuum and then find the nearest node that is just greater than the key's point. When a node is removed only those sections of the continuum that hashed to that node are redistributed across the remaining nodes, rather than rehashing the entire keyspace amongst the remaining nodes.

Example:

  hosts := []string{"host1", "host2"}
  c := ketama.Make(hosts)
  host := c.GetHost('my key')
*/
package ketama
