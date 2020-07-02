package vsn

import (
	"fmt"

	"github.com/fatih/color"
)

const banner = `
____ ____ ___  _  _ ____ ____ ____ ____ ____ ____ ___
|__, [__] |--' |--| |=== |--< |___ |--< |--| |---  | 

The programs included with Gophercraft are free software;
the exact distribution terms for each program are described in LICENSE.

`

func PrintBanner() {
	color.Set(color.FgCyan)
	fmt.Println(banner)
	color.Unset()
}
