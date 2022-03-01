package external

import "github.com/urfave/cli/v2"

type MyError struct {
	cli.Author
}
