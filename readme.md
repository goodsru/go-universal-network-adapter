# go-universal-network-adapter

[![GitHub license](https://img.shields.io/badge/License-MIT-lightgrey.svg)](https://github.com/avito-tech/Marshroute/blob/master/LICENSE) ![GitHub release](https://img.shields.io/badge/Version-1.0.0-brightgreen.svg)

`go-universal-network-adapter` is a Go library for managing file downloads using different network protocols. 
It does a good job helping you build applications, in which remote files are being browsed, downloaded and processed. 
Universal Network Adapter offers an elegant abstraction over major network protocols and as such, it could be easily extended or just plugged into your project.

## Install

`go get -u github.com/konart/go-universal-network-adapter`


## Features

* **compatible with http/https/ftp/ftps/sftp** - most popular protocols to work with file servers
* **tls support** - allows connect to server with certificate (not only basic login/password)


## Examples

See [examples](https://github.com/konart/go-universal-network-adapter/tree/master/examples) for a variety of examples.


**Http:**

```go
package main

import (
	"fmt"
	"github.com/konart/go-universal-network-adapter/models"
	"github.com/konart/go-universal-network-adapter/services"
	"io/ioutil"
	"time"
)

func main() {
    adapter := services.NewUniversalNetworkAdapter()
    remoteFile, err := models.NewRemoteFile(models.NewDestination("http://lorempixel.com/400/200/", nil, nil))
    if err != nil {
        panic(err)
    }
    content, err := adapter.Download(remoteFile)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(content.Name)
    fmt.Println(content.Path)
    buf, err := ioutil.ReadAll(content.Blob)
    fmt.Println(len(buf))
}

```

**Ftp:**

In case you don't have production ftp server, then for testing ftp downloading first you need to up ftp server on localhost, as it's implemented in this
[example](https://github.com/konart/go-universal-network-adapter/tree/master/examples/example_ftp.go).
```go
package main

import (
	"fmt"
	"github.com/konart/go-universal-network-adapter/models"
	"github.com/konart/go-universal-network-adapter/services"
	"io/ioutil"
)

func main() {
	adapter := services.NewUniversalNetworkAdapter()

	remoteFile, err := models.NewRemoteFile(models.NewDestination("ftp://localhost:21/test1.txt", &models.Credentials{
		User:     "user",
		Password: "pass",
	}, nil))

	if err != nil {
		panic(err)
	}
	content, err := adapter.Download(remoteFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(content.Name)
	fmt.Println(content.Path)
	buf, err := ioutil.ReadAll(content.Blob)
	fmt.Println(string(buf))
}

```
**Sftp:**

In case you don't have some production sftp server, then for testing sftp downloading first you need to up sftp server 
on localhost, as it's implemented in this [example](https://github.com/konart/go-universal-network-adapter/tree/master/examples/example_sftp.go).

```go
package main

import (
	"fmt"
	"io/ioutil"
	"github.com/konart/go-universal-network-adapter/models"
	"github.com/konart/go-universal-network-adapter/services"
)

func main() {
	adapter := services.NewUniversalNetworkAdapter()

	remoteFile, err := models.NewRemoteFile(models.NewDestination("sftp://host.com:22/folder/file.json", &models.Credentials{
		User:     "Admin",
		Password: "Qwerty",
	}, nil))

	if err != nil {
		panic(err)
	}
	content, err := adapter.Download(remoteFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(content.Name)
	fmt.Println(content.Path)
	buf, err := ioutil.ReadAll(content.Blob)
	fmt.Println(string(buf))
}

```

## Credits

* Contributions: [@MilovanovIvan](https://github.com/MilovanovIvan), [@eekrupin](https://github.com/eekrupin), [@dmitrymatviets](https://github.com/dmitrymatviets)

## License
```
The MIT License (MIT)

Copyright (c) 2019 LLC Marketplace

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
