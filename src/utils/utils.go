package utils

import "strings"

func ParseArgs(args []string) map[string][]string {
	var flags []string
	var nonFlags []string
	result := make(map[string][]string)
	i := 0

	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				flags = append(flags, parts[0], parts[1])
			} else {
				flags = append(flags, arg)

				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					flags = append(flags, args[i])
				}
			}
		} else {
			nonFlags = append(nonFlags, arg)
		}
		i++
	}

	result["flags"] = flags
	result["nonFlags"] = nonFlags
	return result
}
