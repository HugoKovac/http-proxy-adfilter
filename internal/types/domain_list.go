package types



type DomainList struct {
	Name string			`json:"name"`
	Description string	`json:"description"`
	List []string		`json:"list"`
}
