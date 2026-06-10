package system

type ImageJson struct {
	ModelName     string      `json:"model"`
	ModelRevision int         `json:"revision"`
	Architecture  string      `json:"architecture"`
	Snaps         []ImageSnap `json:"snaps"`
}

type ImageSnap struct {
	SnapName     string `json:"name"`
	SnapRevision int    `json:"revision"`
	DownloadUrl  string `json:"-"` // Just for internal convenience
	Assertion    string `json:"-"` // Just for internal convenience
}
