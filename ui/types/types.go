package types

// holds common data for all templates
type TemplateData struct {
	CurrentYear     int
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}