package utils

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	version = "1.0.0"
)

func PrintBanner() {
	banner := `
  ┌─┐┬  ┌─┐┌┬┐┌─┐
  └─┐│  ├─┤ │ ├┤ 
  └─┘┴─┘┴ ┴ ┴ └─┘
`

	fmt.Println(color.CyanString(banner))
	fmt.Println(color.WhiteString(
		fmt.Sprintf("        Slate · Terminal Presentation Tool v%s", version),
	))
	fmt.Println(color.YellowString(
		"        https://github.com/rivetron/slate",
	))
	fmt.Println()
}
