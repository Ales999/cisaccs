# cisaccs

Group Cisco access with group configs

Example use:

Use exampe files from 'example' dir.

```go
package main

import (
 "github.com/ales999/cisaccs"
)


func main() {

 portSsh : = 22
 cisFileName := "/etc/cisco/cis.yaml"
 pwdFileName := "/etc/cisco/passw.json"

 acc := cisaccs.NewCisAccount(cisFileName, pwdFileName)
 err := acc.OneCisExecuteSsh("gns3-r2", portSsh, []string{"sh arp"})
 if err != nil {
  panic(err)
 }

}
```

PS Using modifed version [netrasp](https://github.com/mrzack99s/netrasp) library with Apache 2.0 license.

