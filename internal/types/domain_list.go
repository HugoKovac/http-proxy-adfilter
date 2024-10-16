package types



type DomainList struct {
	Name string			`json:"name"`
	Description string	`json:"description"`
	List []string		`json:"list"`
}

type CategoryList struct {
	CategoryName string `json:category_name`
	Description string `json:description`
}
