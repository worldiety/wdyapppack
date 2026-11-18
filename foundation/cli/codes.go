package cli

import "github.com/worldiety/wdyapppack/foundation/problem"

// Codes documents the messages of the command line layer itself.
func Codes() []problem.Doc {
	return []problem.Doc{
		{Code: "WAP-COMMAND-UNKNOWN", Summary: "the first word is not a command; the closest command is suggested", See: "help"},
		{Code: "WAP-OPTION-UNKNOWN", Summary: "a command does not know an option; the closest option is suggested", See: "help"},
		{Code: "WAP-OPTION-VALUE", Summary: "an option lacks its value, or its value cannot be used", See: "help"},
		{Code: "WAP-ARGUMENT-MISSING", Summary: "a command needs an argument that was not given", See: "help"},
		{Code: "WAP-ARGUMENT-EXTRA", Summary: "more arguments were given than the command takes, often an option without its dashes or a path with a space", See: "help"},
		{Code: "WAP-TOPIC-UNKNOWN", Summary: "help was asked about a command or topic that does not exist", See: "help"},
		{Code: "WAP-INTERNAL", Summary: "an error nobody anticipated; it carries its original text", See: "errors"},
	}
}
