# vars - a package for atomically capturing counters and other state.

## Overview

When running applications it helps, for debugging purposes, to support
counting events in a standardized way. There are some elaborate
packages out there for doing this, but this package is a bare bones
version that I have found useful.

## Samples

To set things up, you request some new metrics:
```
m := vars.New()
```
Then you can add to counters like this:
```
m.Add("my-counter", 2)
```
or simply set a metric:
```
m.Set("a-record", "green")
```
You can obtain the current value of a metric with:
```
value := m.Get("my-counter")
```
which simply returns `nil` if the requested metric does not exist.

## License info

The `vars` package is distributed with the same BSD 3-clause license
as that used by [golang](https://golang.org/LICENSE) itself.

## Reporting bugs and feature requests

The `vars` package was developed purely out of self-interest to help
debug other programs and packages. Should you find a bug or want to
suggest a feature addition, please use the [bug
tracker](https://github.com/tinkerator/vars/issues).
