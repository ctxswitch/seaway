// Copyright 2024 Seaway Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package console

import (
	"fmt"
	"os"
	"strings"

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
	Ex         = "✘ "

	Line = "─"
)

//
// TODO: I'm not going to bother with documenting this one.  It needs to change and was just
// a quick hack to get something working.
//

func Section(format string, a ...any) {
	line := fmt.Sprintf(format, a...)
	sep := strings.Repeat(Line, len(line))
	sec := chalk.Bold.TextStyle(chalk.Green.String() + line + "\n" + chalk.Reset.String())
	sec += chalk.Dim.TextStyle(chalk.White.String() + sep + "\n" + chalk.Reset.String())
	fmt.Print("\n" + sec)
}

func Info(format string, a ...any) {
	format = chalk.Dim.TextStyle(FishEye) + chalk.Bold.TextStyle(chalk.Blue.String()+format+"\n"+chalk.Reset.String())
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
	format = chalk.Red.String() + Ex + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
	os.Exit(1)
}

func Success(format string, a ...any) {
	format = chalk.Dim.TextStyle(BullsEye) + chalk.Bold.TextStyle(chalk.Green.String()+format+"\n"+chalk.Reset.String())
	fmt.Printf(format, a...)
}

func ListNotice(format string, a ...any) {
	format = "  " + chalk.Dim.TextStyle(Circle) + chalk.Cyan.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}

func ListWarning(format string, a ...any) {
	format = "  " + chalk.Dim.TextStyle(Circle) + chalk.Yellow.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}

func ListError(format string, a ...any) {
	format = "  " + chalk.Dim.TextStyle(Circle) + chalk.Red.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}

func ListFailed(format string, a ...any) {
	format = "  " + chalk.Dim.TextStyle(BullsEye) + chalk.Red.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}

func ListSuccess(format string, a ...any) {
	format = "  " + chalk.Dim.TextStyle(BullsEye) + chalk.Green.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}

func Newline() {
	fmt.Println()
}

func Created(format string, a ...any) {
	bullet := "  " + chalk.Green.String() + CheckBox + chalk.Reset.String()
	format = bullet + format + "\n"
	fmt.Printf(format, a...)
}

func Updated(format string, a ...any) {
	bullet := "  " + chalk.Yellow.String() + CheckBox + chalk.Reset.String()
	format = bullet + format + "\n"
	fmt.Printf(format, a...)
}

func Unchanged(format string, a ...any) {
	bullet := "  " + chalk.Blue.String() + CheckBox + chalk.Reset.String()
	format = bullet + chalk.Dim.TextStyle(format+"\n")
	fmt.Printf(format, a...)
}

func Waiting(format string, a ...any) {
	format = "  " + chalk.Green.String() + CheckBox + chalk.Reset.String() + chalk.Inverse.String() + format + "\n" + chalk.Reset.String()
	fmt.Printf(format, a...)
}
