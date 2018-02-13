package main

import (
	"os"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

// func TestColor(t *testing.T) {
// color.Set(color.FgYellow)
// logit.Info(" Sent %d messages to ES Bulk Processor", 72)
// color.Unset()

// fmt.Printf("This is a %s and this is %s.\n", yellow("warning"), red("error"))

// fmt.Printf("This is a %s and this is %s.\n", yellow("warning"), red("error"))
// 	logit.Info(" %s messages processed from %s since last reset", yellow(2), green("blah"))
// 	logit.Info(" Current queue length for %s is %s", green("blah"), red(6))
// }

// func TestToString(t *testing.T) {
// 	s := strconv.FormatInt(42, 10)
// 	fmt.Println(s)
// }
