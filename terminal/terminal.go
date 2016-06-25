package terminal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"golang.org/x/crypto/ssh/terminal"
)

// ##################################################################
//
// Prompts
//
// ##################################################################

type boxMessage struct {
	Title   string
	Message string
	Blank   string
}

func PrintAnsi(templ string, data interface{}) {
	ansiTemplate := template.FuncMap{
		"ansi": AnsiCode,
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Funcs(ansiTemplate).Parse(templ))
	err := t.Execute(w, data)
	if err != nil {
		panic(err)
	}
	w.Flush()
}

// AnsiCode outputs the ansi codes for changing terminal colors and behaviors.
func AnsiCode(code string) string {
	var ansiCode int
	switch code {
	case "reset":
		ansiCode = 0
	case "bright":
		ansiCode = 1
	case "dim":
		ansiCode = 2
	case "underscore":
		ansiCode = 4
	case "blink":
		ansiCode = 5
	case "reverse":
		ansiCode = 7
	case "hidden":
		ansiCode = 8
	case "fgblack":
		ansiCode = 30
	case "fgred":
		ansiCode = 31
	case "fggreen":
		ansiCode = 32
	case "fgyellow":
		ansiCode = 33
	case "fgblue":
		ansiCode = 34
	case "fgmagenta":
		ansiCode = 35
	case "fgcyan":
		ansiCode = 36
	case "fgwhite":
		ansiCode = 37
	case "bgblack":
		ansiCode = 40
	case "bgred":
		ansiCode = 41
	case "bggreen":
		ansiCode = 42
	case "bgyellow":
		ansiCode = 43
	case "bgblue":
		ansiCode = 44
	case "bgmagenta":
		ansiCode = 45
	case "bgcyan":
		ansiCode = 46
	case "bgwhite":
		ansiCode = 47
	}
	output := fmt.Sprintf("\033[%dm", ansiCode)
	return output
}

var ErrorMessageTemplate = `
{{ ansi "fgwhite"}}{{ ansi "bgred"}}{{.Blank}}{{ ansi ""}}
{{ ansi "fgwhite"}}{{ ansi "bgred"}}{{.Title}}{{ ansi ""}}
{{ ansi "fgwhite"}}{{ ansi "bgred"}}{{.Blank}}{{ ansi ""}}
{{ ansi "fgwhite"}}{{ ansi "bgred"}}{{.Message}}{{ ansi ""}}
{{ ansi "fgwhite"}}{{ ansi "bgred"}}{{.Blank}}{{ ansi ""}}
`

/*
var BoxPromptTemplate = `
{{ ansi "reverse"}}{{.Blank}}{{ ansi ""}}
{{ ansi "reverse"}}{{.Title}}{{ ansi ""}}
{{ ansi "reverse"}}{{.Blank}}{{ ansi ""}}
{{ ansi "reverse"}}{{.Message}}{{ ansi ""}}
{{ ansi "reverse"}}{{.Blank}}{{ ansi ""}}
`

var InformationTemplate = `
{{ ansi "reverse"}}{{.}}{{ ansi ""}}
`
*/

var BoxPromptTemplate = `
{{ ansi "fgblack"}}{{ ansi "bgcyan"}}{{.Blank}}{{ ansi ""}}
{{ ansi "fgblack"}}{{ ansi "bgcyan"}}{{.Title}}{{ ansi ""}}
{{ ansi "fgblack"}}{{ ansi "bgcyan"}}{{.Blank}}{{ ansi ""}}
{{ ansi "fgblack"}}{{ ansi "bgcyan"}}{{.Message}}{{ ansi ""}}
{{ ansi "fgblack"}}{{ ansi "bgcyan"}}{{.Blank}}{{ ansi ""}}
`

var InformationTemplate = `
{{ ansi "fgblack"}}{{ ansi "bgcyan"}}{{.}}{{ ansi ""}}
`
var ErrorLineTemplate = `
{{ ansi "fgwhite"}}{{ ansi "bgred"}}{{.}}{{ ansi ""}}
`

// Error Message
func ShowErrorMessage(title string, message string) {
	boxMessage := prepMessage(title, message)
	PrintAnsi(ErrorMessageTemplate, boxMessage)
}

// Input Prompt Bool
func BoxPromptBool(title string, message string) bool {

	boxMessage := prepMessage(title, message)
	PrintAnsi(BoxPromptTemplate, boxMessage)

	return askForConfirmation()
}

// Input Prompt String
func BoxPromptString(title string, message string) string {

	boxMessage := prepMessage(title, message)
	PrintAnsi(BoxPromptTemplate, boxMessage)

	return askForString()
}

func PromptString(message string) string {
	Information(message)
	return askForString()
}

func PromptPassword(message string) string {
	Information(message)
	return askForPassword()
}

func PromptInt(message string, max int) int {
	Information(message)
	return askForInt(max)
}

func PromptBool(message string) bool {
	Information(message)
	return askForConfirmation()
}

func Information(message string) {
	message = strings.Replace(message, "\n", " ", -1)
	message = strings.Replace(message, "\t", " ", -1)
	message = padStringRight(message, 100)
	PrintAnsi(InformationTemplate, message)
}

func ErrorLine(message string) {
	message = strings.Replace(message, "\n", " ", -1)
	message = strings.Replace(message, "\t", " ", -1)
	message = padStringRight(message, 100)
	PrintAnsi(ErrorLineTemplate, message)
}

// formats and prints a title and message in a template block
func prepMessage(title string, message string) boxMessage {

	message = strings.Replace(message, "\n", " ", -1)
	message = strings.Replace(message, "\t", " ", -1)

	title = fmt.Sprintf("[%s]", title)

	// Figure out how wide of a notification we are going to be building
	titleWidth := len(title)
	msgWidth := len(message)
	var totalWidth int
	// TODO make this use math.Max or something? idk
	if titleWidth > msgWidth {
		totalWidth = titleWidth + 4
	} else {
		totalWidth = msgWidth + 4
	}

	totalWidth += (totalWidth % 2)

	// Pad our strings until they are centered at the same width
	title = padStringCenter(title, totalWidth)
	message = padStringCenter(message, totalWidth)
	blank := strings.Repeat(" ", totalWidth)

	/*
		title = padStringCenter(title, 100)
		message = padStringCenter(message, 100)
		blank := strings.Repeat(" ", 100)
	*/

	return boxMessage{Title: title, Message: message, Blank: blank}
}

// Padding helpers
func padStringCenter(s string, w int) string {
	if len(s)%2 != 0 {
		s += " "
	}

	padding := strings.Repeat(" ", (w-len(s)%w)/2)
	t := []string{padding, s, padding}
	padded := strings.Join(t, "")
	return padded
}

func padStringRight(s string, w int) string {

	if w < len(s) {
		w = len(s) + 1
	}

	padding := strings.Repeat(" ", (w - len(s)))
	t := []string{"  ", s, padding}
	padded := strings.Join(t, "")

	return padded
}

func askForString() string {

	var resp string

	fmt.Print("> ")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		resp = scanner.Text()

		// Check for errors during `Scan`. End of file is
		// expected and not reported by `Scan` as an error.
		if err := scanner.Err(); err != nil || len(resp) == 0 {
			Information("Please enter a value and then press enter:")
			return askForString()
		} else {
			break
		}
	}

	return resp
}

func askForPassword() string {

	state, err := terminal.MakeRaw(0)
	if err != nil {
		return ""
	}

	defer terminal.Restore(0, state)

	term := terminal.NewTerminal(os.Stdout, "")

	bytePassword, err := term.ReadPassword("> ")
	if err != nil {
		return ""
	}

	fmt.Println(strings.Repeat("*", len(bytePassword)))

	return strings.TrimSpace(string(bytePassword))
}

func askForInt(max int) int {

	fmt.Print("> ")

	var i int

	_, err := fmt.Scan(&i)

	if err == nil && i > 0 && i <= max {
		return i
	}

	if i == 0 {
		var discard string
		fmt.Scanln(&discard) // eww
	}

	Information(fmt.Sprintf("Please enter a number between 1 and %d, then press enter:", max))
	return askForInt(max)
}

// From: https://gist.github.com/albrow/5882501
func askForConfirmation() bool {

	fmt.Print("> ")
	var response string
	fmt.Scanln(&response)

	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		Information("Please type 'yes' or 'no' and then press enter:")
		return askForConfirmation()
	}
}

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true if slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}
