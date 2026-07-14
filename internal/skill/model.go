package skill

type Root struct {
	Path      string
	Namespace string
}

type Manifest struct {
	Version       int
	ID            string
	Phrases       []string
	Extensions    []string
	Environments  []string
	Requires      []string
	ComposesWith  []string
	ConflictsWith []string
	Supersedes    []string
}

type Skill struct {
	ID          string
	Namespace   string
	Name        string
	Description string
	SkillMD     string
	Directory   string
	BodyHash    string
	Manifest    Manifest
}

type Diagnostic struct {
	Path     string `json:"path"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

func hasErrors(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == "error" {
			return true
		}
	}
	return false
}
