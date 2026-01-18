package buildinfo

import "strings"

const (
	devVersion = "dev"
	unknown    = "unknown"
)

// これらの変数の値は本番ビルド時に-ldflagsによって設定されるメタ情報
var (
	version   = devVersion
	commitID  = unknown
	buildTime = unknown
	goBuild   = unknown
)

func Version() string {
	return version
}

func VersionWithPrefix() string {
	if version == devVersion {
		return version
	}
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func CommitID() string {
	return commitID
}

func ShortCommitID() string {
	if commitID == unknown {
		return unknown
	}
	const n = 7
	if len(commitID) <= n {
		return commitID
	}
	return commitID[:n]
}

func BuildTime() string {
	return buildTime
}

func GoBuild() string {
	return goBuild
}
