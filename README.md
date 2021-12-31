vso-hash
========

Golang implementation of the BuildXL paged hash function

See https://github.com/microsoft/BuildXL/blob/master/Documentation/Specs/PagedHash.md for
more information about the function in general.


What are the benefits?
----------------------

VSO-Hash is essentially a sharded SHA256 which can be relatively easily parallelised.
This can reap significant benefits on modern machines:

```
goos: linux
goarch: amd64
pkg: github.com/peterebden/vso-hash
cpu: AMD Ryzen 9 5900X 12-Core Processor
BenchmarkVSOHash/SHA256-24         	       1	3347418236 ns/op	 611.82 MB/s
BenchmarkVSOHash/VSOParallel1-24   	       1	3361708658 ns/op	 609.21 MB/s
BenchmarkVSOHash/VSOParallel2-24   	       1	2140195088 ns/op	 956.92 MB/s
BenchmarkVSOHash/VSOParallel4-24   	       1	1021680305 ns/op	2004.54 MB/s
BenchmarkVSOHash/VSOParallel8-24   	       2	 588285173 ns/op	3481.31 MB/s
BenchmarkVSOHash/VSOParallel16-24  	       3	 450288525 ns/op	4548.20 MB/s
BenchmarkVSOHash/VSOParallel24-24  	       3	 425042098 ns/op	4818.35 MB/s
PASS
```

It is one of the hash options in the [Remote Execution API](https://github.com/bazelbuild/remote-apis).
