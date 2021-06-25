package main

import (
	"encoding/json"
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

	o, err := internal.LoadOrganization(c.OrganizationID)
	if err != nil {
		fmt.Println(err)
		return
	}

	o.LoadTeams()

	j, err := json.Marshal(o)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(j))
}