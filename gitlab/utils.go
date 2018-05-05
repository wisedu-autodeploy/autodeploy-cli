package gitlab

import (
	"strconv"
	"strings"
)

// addTagVersion can major/minor/patch a tag version by what t is(major/minor/patch).
func addTagVersion(tag string, t string) string {
	splices := strings.Split(tag, "v")
	versionSplices := strings.Split(splices[1], ".")

	major := versionSplices[len(versionSplices)-3]
	minor := versionSplices[len(versionSplices)-2]
	patch := versionSplices[len(versionSplices)-1]

	switch t {
	case "major":
		num, _ := strconv.Atoi(major)
		major = strconv.Itoa(num + 1)
	case "minor":
		num, _ := strconv.Atoi(minor)
		minor = strconv.Itoa(num + 1)
	case "patch":
		num, _ := strconv.Atoi(patch)
		patch = strconv.Itoa(num + 1)
	}

	versionSplices[len(versionSplices)-3] = major
	versionSplices[len(versionSplices)-2] = minor
	versionSplices[len(versionSplices)-1] = patch

	newVersion := strings.Join(versionSplices, ".")
	newTag := strings.Join([]string{splices[0], newVersion}, "v")

	return newTag
}
