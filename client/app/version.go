package app

import (
	_ "embed"

	"decred.org/dcrdex/dex/version"
)

/*
Note for maintainers:

The expected process for setting the version in releases is as follows:
- Create a release branch of the form 'release-vMAJOR.MINOR'
- Modify the Version variable below on that branch to:
	- Remove the pre-release portion
	- Set the build metadata to 'release.local'
	- Example: 'Version = "0.5.0+release.local"'
- Update the Version variable below on the master branch to the next
	expected version while retaining a pre-release of 'pre'

These steps ensure that building from source produces versions that are
distinct from reproducible builds that override the Version via linker
flags.

Version is the application version per the semantic versioning 2.0.0 spec
(https://semver.org/).

It is defined as a variable so it can be overridden during the build
process with:
'-ldflags "-X main.Version=fullsemver"'
if needed.

It MUST be a full semantic version per the semantic versioning spec or
the package will panic at runtime.  Of particular note is the pre-release
and build metadata portions MUST only contain characters from
semanticAlphabet.
NOTE: The Version string is overridden on init.

The version is set in a text file and go:embed-ed, so that packaging scripts can
easily read it and stay in sync.
*/

//go:embed VERSION
var Version string

func init() {
	Version = version.Parse(Version)
}
