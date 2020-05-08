package colors

import "fmt"

type Code int

const (
	CodeRed         Code = 31
	CodeGreen       Code = 32
	CodeYellow      Code = 33
	CodeBlue        Code = 34
	CodePurple      Code = 35
	CodeCyan        Code = 36
	CodeGray        Code = 37
	CodeRedLight    Code = 91
	CodeGreenLight  Code = 92
	CodeYellowLight Code = 93
	CodeBlueLight   Code = 94
	CodePurpleLight Code = 95
	CodeCyanLight   Code = 96
	CodeWhite       Code = 97
	CodeReset       Code = 0
	CodeBold        Code = 1
)

func (c Code) String() string {
	return fmt.Sprintf("\u001B[%dm", c)
}

type Color string

const (
	Red         Color = "red"
	Green       Color = "green"
	Yellow      Color = "yellow"
	Blue        Color = "blue"
	Purple      Color = "purple"
	Cyan        Color = "cyan"
	Gray        Color = "gray"
	RedLight    Color = "red_light"
	GreenLight  Color = "green_light"
	YellowLight Color = "yellow_light"
	BlueLight   Color = "blue_light"
	PurpleLight Color = "purple_light"
	CyanLight   Color = "cyan_light"
	White       Color = "white"
	Reset       Color = "reset"
	Bold        Color = "bold"
)

func (c Color) String() string {
	switch c {
	case Red:
		return CodeRed.String()
	case Green:
		return CodeGreen.String()
	case Yellow:
		return CodeYellow.String()
	case Blue:
		return CodeBlue.String()
	case Purple:
		return CodePurple.String()
	case Cyan:
		return CodeCyan.String()
	case Gray:
		return CodeGray.String()
	case RedLight:
		return CodeRedLight.String()
	case GreenLight:
		return CodeGreenLight.String()
	case YellowLight:
		return CodeYellowLight.String()
	case BlueLight:
		return CodeBlueLight.String()
	case PurpleLight:
		return CodePurpleLight.String()
	case CyanLight:
		return CodeCyanLight.String()
	case White:
		return CodeWhite.String()
	case Reset:
		return CodeReset.String()
	case Bold:
		return CodeBold.String()
	}

	return ""
}
