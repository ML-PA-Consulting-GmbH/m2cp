package console

type Color int

const (
	Reset Color = iota
	Red
	Green
	Yellow
	Blue
	Purple
	Cyan
	White
)

var colorCodes = map[Color]string{
	Reset:  "\033[0m",
	Red:    "\033[31m",
	Green:  "\033[32m",
	Yellow: "\033[33m",
	Blue:   "\033[34m",
	Purple: "\033[35m",
	Cyan:   "\033[36m",
	White:  "\033[37m",
}

func Colorize(color Color, text string) string {
	return colorCodes[color] + text + colorCodes[Reset]
}
