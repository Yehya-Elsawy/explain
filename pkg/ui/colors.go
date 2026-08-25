package ui

import (
	"os"
)

var (
	NoColor = false

	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BoldRed     = "\033[1;31m"
	BoldGreen   = "\033[1;32m"
	BoldYellow  = "\033[1;33m"
	BoldCyan    = "\033[1;36m"
	BoldWhite   = "\033[1;37m"
	BoldMagenta = "\033[1;35m"

	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"
	BgCyan   = "\033[46m"
)

func InitColors(disableColor bool) {
	if disableColor || os.Getenv("NO_COLOR") != "" {
		NoColor = true
		Reset = ""
		Bold = ""
		Dim = ""
		Italic = ""
		Underline = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Magenta = ""
		Cyan = ""
		White = ""
		BoldRed = ""
		BoldGreen = ""
		BoldYellow = ""
		BoldCyan = ""
		BoldWhite = ""
		BoldMagenta = ""
		BgRed = ""
		BgGreen = ""
		BgYellow = ""
		BgBlue = ""
		BgCyan = ""
	}
}

func Colorize(color, text string) string {
	if NoColor {
		return text
	}
	return color + text + Reset
}
