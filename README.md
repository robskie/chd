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

// Create a new builder
b := chd.NewBuilder()

// Add your items to the builder
for key, value := range data {
  b.Add(key, value)
}

// Build the map
m := b.Build(nil)

// Do something useful with the map
```

To iterate over the key-value pairs, you can get an iterator by calling the
Map.Iterator method.

```go
// Iterate over key-value pairs
for it := m.Iterator(); it != nil; it = it.Next() {
  key := it.Key()
  value := it.Value()

  // Do something useful with key/value
}


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
BenchmarkBuild10KKeys         20       93784414 ns/op
BenchmarkBuild100KKeys         1     1399059965 ns/op
BenchmarkBuild1MKeys           1    34712674432 ns/op
BenchmarkMapGetIntArray  5000000            289 ns/op
BenchmarkMapGetFibArray  2000000            658 ns/op
```
