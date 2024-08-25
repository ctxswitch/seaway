package console

import (
	"fmt"
	"os"

	"github.com/ttacon/chalk"
)

const (
	Circle     = "◦ "
	RightArrow = "⇨ "
	Square     = "■ "
	FishEye    = "◉ "
	BullsEye   = "◎ "
	CheckBox   = "☑ "
	ExBox      = "☒ "
	EmptyBox   = "☐ "
	CheckMark  = "✔ "
)

func Info(format string, a ...any) {
	format = chalk.Dim.TextStyle(Square) + chalk.Bold.TextStyle(chalk.Blue.String()+format+"\n"+chalk.Reset.String())
	fmt.Printf(format, a...)
}

func Error(format string, a ...any) {
	format = chalk.Red.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}

func Warn(format string, a ...any) {
	format = chalk.Yellow.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}

func ListItem(format string, a ...any) {
	format = "  " + chalk.Dim.TextStyle(RightArrow) + chalk.White.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(chalk.Italic.TextStyle(format), a...)
}

func Fatal(format string, a ...any) {
	format = chalk.Red.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
	os.Exit(1)
}

func Success(format string, a ...any) {
	format = chalk.Dim.TextStyle(BullsEye) + chalk.Bold.TextStyle(chalk.Green.String()+format+"\n"+chalk.Reset.String())
	fmt.Printf(format, a...)
}

func Notice(format string, a ...any) {
	format = "  " + chalk.Dim.TextStyle(Circle) + chalk.Cyan.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}

func Newline() {
	fmt.Println()
}

func Created(format string, a ...any) {
	format = "  " + chalk.Green.String() + CheckMark + chalk.Reset.String() + chalk.Inverse.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}

func Updated(format string, a ...any) {
	format = "  " + chalk.Yellow.String() + CheckMark + chalk.Reset.String() + chalk.Inverse.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}

func Unchanged(format string, a ...any) {
	format = "  " + chalk.Blue.String() + CheckMark + chalk.Reset.String() + chalk.Inverse.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}
