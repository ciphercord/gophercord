> [!CAUTION]
> Anyone can interact with the CipherCord API. Please sanitize responses to avoid errors and possible hijacking attempts.

# GopherCord [![MIT License](https://img.shields.io/badge/License-MIT-a10b31)](https://github.com/ciphercord/gophercord/blob/main/LICENSE)

**GopherCord** is a set of Go packages designed to aid in interacting with the CipherCord API from Go.

## Example
```go
// simple message net
package main

import (
	"fmt"
	"log"

	ccbot "github.com/ciphercord/gophercord/bot"
	ccmsg "github.com/ciphercord/gophercord/message"
)

func main() {
	if err := ccbot.Init(); err != nil {
		log.Fatal(err)
	}

	for {
		data := <-ccbot.Messages

		umsg, err := ccmsg.Unpackage(data, "MyPrivateKey")
		if err == ccmsg.ErrUnmatched {
			continue
		} else if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s: %s\n", umsg.Author, umsg.Content)
	}
}
```

