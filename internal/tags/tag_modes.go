package tags

import "strings"

type TagMode int

const (
	TagModeNormal   TagMode = iota // merge policy tags + AEGIS.TAG
	TagModeOverride                // AEGIS.TAG_ONLY — ignore policy tags
	TagModeSkip                    // AEGIS.NOTAG — skip tagging entirely
)

// TODO: handle cases like aegis commands are not last
// parseTagArgs scans SET args for AEGIS.TAG, AEGIS.TAG_ONLY, AEGIS.NOTAG
func parseTagArgs(args []string) (TagMode, []string) {
	// parse all args and check for tag modifiers
	for i := 0; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "AEGIS.NOTAG":
			return TagModeSkip, nil
		case "AEGIS.TAG_ONLY":
			return TagModeOverride, args[i+1:]
		case "AEGIS.TAG":
			return TagModeNormal, args[i+1:]
		}
	}
	return TagModeNormal, nil
}
