package snapdseed

type SeedTemplate struct {
	Store SeedStore  `yaml:"store"`
	Model SeedModel  `yaml:"model"`
	Snaps []SeedSnap `yaml:"snaps"`
}

type SeedStore struct {
	URL string `yaml:"url"`
}

type SeedModel struct {
	Name     string `yaml:"name"`
	Revision int    `yaml:"revision"`
}

type SeedSnap struct {
	Name                     string `yaml:"name"`
	Arch                     string `yaml:"arch"`
	Version                  string `yaml:"version,omitempty"`
	Revision                 int    `yaml:"revision,omitempty"`
	snapDeclarationAssertion string
	snapRevisionAssertion    string
}
