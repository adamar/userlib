
# User library

Create linux system users without shelling out to `useradd`


## Quick start

```
package main

import (
	"fmt"
	"github.com/adamar/userlib"
	)


func main() {


	// Set some basic attributes for the new user
	u := &User{Username: "john", GroupMemberships: []string{"games"}, Shell: "/bin/bash"}

	// Create the user
	u.AddUser()

	// Print out details about the new user
	fmt.Printf("%+v\n", u)

}

```


## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| `Username` | The system username  | `string` | - | yes |
| `Uid` | The users ID | `string` | - | no |
| `Gid` | The group ID | `string` | - | no |
| `Groupname` | The name of the users main group | `string` | - | no |
| `Homedir` | The path to the users home directory | `string` | - | no |
| `Shell` | The shell assigned to the user | `string` | `/sbin/nologin` | no |
| `GroupMemberships` | The extra groups delegated to the user | `list(string)`| - | no |



