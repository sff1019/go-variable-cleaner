# varcleaner

variable-cleaner solves issues such as:
- Detect variables that does not needed to be declared (e.g. only used once)
- Detect redundant constants and warn users to replace it with a variable

This tool aims to make codes readable, and more scalable.

## Example

```
package a 

import "fmt"

func f() {
  a := "Hello"
  fmt.Println(a)
}
```

Warning Message:
```
No need to define these variables: a
```

```
package a 

import "fmt"

func f() {
  fmt.Println("Hello")
  fmt.Println("Hello")
}
```

Warning Message:
```
Used same consts multiple times, replace with variable: \"Hello\"
```

## Install
$ go get github.com/sff1019/varcleaner/cmd/varcleaner

## Usage
$ go vet -vettool=`function` pkgname
