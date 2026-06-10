package suit

type IntermediateFormat struct {
	AuthenticationWrapper []interface{} `json:"authentication-wrapper"`
	Manifest              Manifest      `json:"manifest"`
}

type Manifest struct {
	ManifestVersion        int       `json:"manifest-version"`
	ManifestSequenceNumber int       `json:"manifest-sequence-number"`
	Common                 Common    `json:"common"`
	Install                []Command `json:"install"`
	Validate               []Command `json:"validate"`
}

type Common struct {
	Components     [][]string `json:"components"`
	CommonSequence []Command  `json:"common-sequence"`
}

type Command struct {
	CommandID   string      `json:"command-id"`
	CommandArg  interface{} `json:"command-arg"`
	ComponentID []string    `json:"component-id"`
}

type Digest struct {
	AlgorithmID string `json:"algorithm-id"`
	DigestBytes string `json:"digest-bytes"`
}
