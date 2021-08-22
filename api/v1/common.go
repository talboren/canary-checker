package v1

import (
	"regexp"
	"strings"

	"github.com/flanksource/canary-checker/api/external"
	"github.com/flanksource/kommons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FolderTest struct {
	//MinAge the latest object should be older than defined age
	MinAge *metav1.Duration `yaml:"minAge,omitempty" json:"minAge,omitempty"`
	//MaxAge the latest object should be younger than defined age
	MaxAge *metav1.Duration `yaml:"maxAge,omitempty" json:"maxAge,omitempty"`
	//MinCount the minimum number of files inside the searchPath
	MinCount *int `yaml:"minCount,omitempty" json:"minCount,omitempty"`
	//MinCount the minimum number of files inside the searchPath
	MaxCount *int `yaml:"maxCount,omitempty" json:"maxCount,omitempty"`
}

type JSONCheck struct {
	Path  string `yaml:"path" json:"path"`
	Value string `yaml:"value" json:"value"`
}

type Authentication struct {
	Username kommons.EnvVar `yaml:"username" json:"username"`
	Password kommons.EnvVar `yaml:"password" json:"password"`
}

func (auth Authentication) GetUsername() string {
	return auth.Username.Value
}

func (auth Authentication) GetPassword() string {
	return auth.Password.Value
}

func (auth Authentication) GetDomain() string {
	parts := strings.Split(auth.GetUsername(), "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

type Display struct {
	external.Template `yaml:",inline" json:",inline"`
}

func (d Display) GetDisplayTemplate() external.Template {
	return d.Template
}

type Test struct {
	external.Template `yaml:",inline" json:",inline"`
}

func (t Test) GetTestTemplate() external.Template {
	return t.Template
}

type Templatable struct {
	Test    external.Template `yaml:"test,omitempty" json:"test,omitempty"`
	Display external.Template `yaml:"display,omitempty" json:"display,omitempty"`
}

func (t Templatable) GetTestTemplate() external.Template {
	return t.Test
}

func (t Templatable) GetDisplayTemplate() external.Template {
	return t.Display
}

type Description struct {
	// Description for the check
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Name of the check
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Icon for overwriting default icon on the dashboard
	Icon string `yaml:"icon,omitempty" json:"icon,omitempty"`
}

func (d Description) GetDescription() string {
	return d.Description
}

func (d Description) GetIcon() string {
	return d.Icon
}

// Obfuscate passwords of the form ' password=xxxxx ' from connectionString since
// connectionStrings are used as metric labels and we don't want to leak passwords
// Returns the Connection string with the password replaced by '###'

func sanitizeEndpoints(connection string) string {
	//looking for a substring that starts with a space,
	//'password=', then any non-whitespace characters,
	//until an ending space
	re := regexp.MustCompile(`\spassword=\S*\s`)
	return re.ReplaceAllString(connection, " password=### ")
}
