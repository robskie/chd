# chd

Package chd implements the compress, hash, and displace (CHD) minimal perfect
hash algorithm described in [Hash, displace, and compress][1] by Botelho et al.
It provides a map builder that manages adding of items and map creation. It also
provides a fibonacci array that can be used to further optimize memory usage.

[1]: http://cmph.sourceforge.net/papers/esa09.pdf

## Installation
```sh
go get github.com/robskie/chd
```

## Example

To create a map, you need to first use a map builder which is where you will
add your items. After you have added all your items, you'll need to call Build
from the map builder to create a map. You will need to supply an array that
implements CompactArray interface to this method or nil to use a plain integer
array instead. If you have a large number of items, like more than a hundred
thousand, building the map would take a while (from a few seconds to several
minutes) to finish.

```go
import "github.com/robskie/chd"

// Sample Item structure
type Item struct {
  Key   string
  Value int
}

items := []Item{
  {"ab", 0},
  {"cd", 1},
  {"ef", 2},
  {"gh", 3},
  {"ij", 4},
  {"kl", 5},
}

// Create a builder and add keys to it
builder := NewBuilder()
for _, item := range items {
  builder.Add([]byte(item.Key))
}

// Build the map
m := builder.Build(nil)

// Rearrange items according to its map index
items = append(items, make([]Item, m.Cap()-len(items))...)
for i := 0; i < len(items); {
  idx := m.Get([]byte(items[i].key))

  if i != idx && len(items[i].key) > 0 {
    items[i], items[idx] = items[idx], items[i]
    continue
  }

  i++
}

// Do something useful
```

You can also serialize a map and deserialize it later by using Map.Read and
Map.Write methods. Like the Builder.Build method, you need to pass a succinct
array that implements CompactArray when deserializing. This should be the same
as the one used in building the map.

```go
// Serialize the map
w, _ := os.Create("mymap.dat")
m.Write(w)

// Afterwards, you can deserialize it
r, _ := os.Open("mymap.dat")
nm := chd.NewMap()
nm.Read(r, nil)

// Do something useful with the map
```

## API Reference

Godoc documentation can be found [here][2].

[2]: https://godoc.org/github.com/robskie/chd

## Benchmarks

A Core i5 running at 2.3GHz is used for these benchmarks. You can see here that
Builder.Build's running time is directly proportional to the number of keys
while the Map.Get's execution time is dependent on the speed of the CompactArray
used to create the map.

You can run these benchmarks on your machine by typing this command
```go test github.com/robskie/chd -bench=.*``` in terminal.

```
BenchmarkBuild10KKeys         30       54484991 ns/op
BenchmarkBuild100KKeys         2      754017901 ns/op
BenchmarkBuild1MKeys           1    16064924548 ns/op
BenchmarkMapGetIntArray  5000000            346 ns/op
BenchmarkMapGetFibArray  2000000            711 ns/op
```
