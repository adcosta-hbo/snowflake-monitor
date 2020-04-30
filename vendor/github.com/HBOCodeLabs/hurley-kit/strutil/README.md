# strutil

> a Go package containing string utility functions and types

### Examples

The package's godocs, and the testable examples located in the various
`*_test.go` source files are always the place to find the most up-to-date
examples. However, for a quick illustration of some of the functionality:

#### `Chunk`
`Chunk` takes a slice of strings splits it into one or more slices of size `chunkSize`. If the original slice isn't evenly divisible by `chunkSize`, the final chunk will contain the remaining elements.

Example: 
```
slice := []string{"hello", "world", "foo", "bar", "baz"}
chunks := strutil.Chunk(slice, 2)

// chunks[0] == []string{"hello", "world"}
// chunks[1] == []string{"foo", "bar"}
// chunks[2] == []string{"baz"}
```

#### `ElideString`
`ElideString` extracts `prefixLen` characters from the beginning, and `suffixLen` characters from the end of a string, and adds an ellipsis (...) in between. If either `prefixLen`, `suffixLen`, or the sum of the two is greater than the length of the string, the string is returned unmodified.

Example:
```
elided := strutil.ElideString("1234567890abcdefghij", 4, 7)
// elided == "1234...defghij"
```

#### `RandomHexString`
`RandomHexString` generates a new pseudo-random string, consisting of `len` lowercase hex digits.  The tests contain a benchmark that shows how fast this method is, since it was adapted from a [very comprehensive SO post](http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang)

Example:
```
str := strutil.RandomHexString(8)
// str == "34bf87d1"
```

#### `Queue`
A FIFO queue of strings.

Example:
```
q := strutil.Queue{}
q.Push("foo")
q.Push("bar")

length := q.Len()
// length == 2

#### `Set`
A hash set of strings.

Example:
```
s := strutil.NewSet(0)
s.Add("foo")
s.Add("bar")
s.Add("baz", "bop")

length := len(s)
// length == 4

for e := range s {
    fmt.Printf("elem: %q\n")
}
// Example output (order not guaranteed)
// "bar"
// "bop"
// "baz"
// "foo"
```

### Development

#### Commands
