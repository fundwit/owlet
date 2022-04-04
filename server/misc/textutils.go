package misc

import "regexp"

var splitRegex *regexp.Regexp

func init() {
	splitRegex = regexp.MustCompile(`[\s,]+`)
}

func Split(input string) []string {
	parts := splitRegex.Split(input, -1)

	output := make([]string, 0, len(parts))
	for _, s := range parts {
		if s == "" {
			continue
		}
		output = append(output, s)
	}
	return output
}
