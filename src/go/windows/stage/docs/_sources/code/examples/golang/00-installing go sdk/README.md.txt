# Example: Installing and using Go SDK

The m2cp-sdk for Go is deployed as a debian (.deb) package.

To install the sdk, use:

`$ sudo dpkg -i mlpa-m2cp-sdk-go-<version>.deb`

Now the sdk ist installed at `/usr/local/lib/mlpa/sdk/go/<version>` and can be used in your Go projects.

To import the sdk, you need to edit the _go.mod_ file of your project. Add a requirement for _m2cp_ to the list
of requirements and a a _replace_ rule at the bottom of the file. Both entries must reflect the version number
of the sdk that you installed. 

Example for using SDK version 6.2.2-dev:

```Go
module main

go 1.21

require (
	m2cp v0.0.0-00010101000000-000000000000
	... other requirements of your project ...
)

require (
    ... list of current indirect requirements ...
)

replace m2cp => /usr/local/lib/mlpa/go/sdk/6.2.2-dev
```

Run `go mod tidy` to refresh your imports. Now you're able to use the sdk in your Go project.

Example:

```Go
package main

import (
	"fmt"
	"m2cp"
)

func main() {
	fmt.Println("Hello, m2cp!")
	fmt.Printf("sdk version: %s", m2cp.GetVersion())
}
```
