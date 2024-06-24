# gitignore

This is a small gitignore-matching library which is based on [go-gitignore](github.com/zabawaba99/go-gitignore), adding a little bit of abstractions.

### Installation

```
go get github.com/shravanasati/gap/gitignore
```

### Usage

```go
package main

import (
	"fmt"
	"github.com/shravanasati/gap/gitignore"
)

func main() {
	matcher, err := gitignore.NewGitignoreMatcher().FromFile("/path/to/your/gitignore")
	// you also have .FromReader(io.Reader) and .FromPatterns([]string)
	// these methods can be chained together, each will extend the internal list of patterns
	
	matcher, err := gitignore.NewGitignoreMatcher().
		FromPatterns([]string{"venv/"})
		FromReader(myTCPReader).

	fmt.Println(matcher.IsIgnored("venv/")) // true
}
```