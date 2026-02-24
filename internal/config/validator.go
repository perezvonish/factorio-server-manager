package config

type Tag string

var (
	Required Tag = "required"
)

func validateRequired(name string, tag string, value string) error {
	if tag == "true" && value == "" {
		return NewFieldRequiredError(name)
	}

	return nil
}
