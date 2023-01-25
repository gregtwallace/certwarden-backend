package dns01manual

import "strings"

// scriptWithArgs returns the script and arguments used when calling script
// i.e. /path/to/script.sh, Domain (2nd Level + TLD), RecordName, RecordValue
func scriptWithArgs(script string, resourceName string, resourceContent string) (args []string) {
	// 0 - script name
	args = append(args, script)

	// 1 - Domain (2nd Level + TLD)
	domainParts := strings.Split(resourceName, ".")
	secondAndTLD := domainParts[len(domainParts)-2] + "." + domainParts[len(domainParts)-1]
	args = append(args, secondAndTLD)

	// 2 - RecordName
	args = append(args, resourceName)

	// 3 - RecordValue
	args = append(args, resourceContent)

	return args
}
