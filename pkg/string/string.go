package string

import(
	"strings"
)


func SanitizeImageName(name string) string {
	return strings.NewReplacer(":", "_", "/", "_").Replace(name)
}