package updater

import (
	"errors"
	"strconv"
	"strings"
)

// errors
var (
	errBadVersionNumber = errors.New("version number failed to parse")
)

// VersionInfo contains data about a specific version of the app
type versionInfo struct {
	Channel       Channel `json:"channel"`
	Version       string  `json:"version"`
	ConfigVersion int     `json:"config_version"`
	URL           string  `json:"url"`
}

// releaseTag returns the name of the Github Release Tag
// func (vi *versionInfo) releaseTag() string {
// 	return "v" + string(vi.Channel)
// }

// isNewerThan compares the specified currentVer against vi and returns true
// if vi is a newer version. If either version doesn't properly parse, an error
// is returned
func (vi *versionInfo) isNewerThan(currentVersionString string) (bool, error) {
	viVers := strings.Split(vi.Version, ".")
	currVers := strings.Split(currentVersionString, ".")

	// make sure versions are proper strings
	// should be exactly 3 values in version
	if len(viVers) != 3 || len(currVers) != 3 {
		return false, errBadVersionNumber
	}

	// each value should be an integer
	majorVi, err := strconv.Atoi(viVers[0])
	if err != nil {
		return false, errBadVersionNumber
	}

	majorCurr, err := strconv.Atoi(currVers[0])
	if err != nil {
		return false, errBadVersionNumber
	}

	minorVi, err := strconv.Atoi(viVers[1])
	if err != nil {
		return false, errBadVersionNumber
	}

	minorCurr, err := strconv.Atoi(currVers[1])
	if err != nil {
		return false, errBadVersionNumber
	}

	patchVi, err := strconv.Atoi(viVers[2])
	if err != nil {
		return false, errBadVersionNumber
	}

	patchCurr, err := strconv.Atoi(currVers[2])
	if err != nil {
		return false, errBadVersionNumber
	}

	// compare version numbers
	switch {
	case majorVi > majorCurr:
		return true, nil
	case majorVi < majorCurr:
		return false, nil
	default:
		switch {
		case minorVi > minorCurr:
			return true, nil
		case minorVi < minorCurr:
			return false, nil
		default:
			if patchVi > patchCurr {
				return true, nil
			} else {
				// patch equal or less than
				return false, nil
			}
		}
	}
}
