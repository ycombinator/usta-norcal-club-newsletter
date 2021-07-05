package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal"
)

func main() {
	c := internal.DefaultConfig()

	if len(os.Args) > 1 {
		oID, err := strconv.ParseInt(os.Args[1], 10, 0)
		if err != nil {
			fmt.Println("organization ID must be an integer")
			return
		}

		c.OrganizationID = int(oID)
	}

	n, err := internal.NewNewsletter(c)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := n.Generate(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(n)
}
