package person

type Person struct {
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Age       int      `json:"age"`
	Children  []Person `json:"children,omitempty"`
}

type Request struct {
	Name
	Subscribe string `json:"subscribe,omitempty"`
}

type Name struct {
	First string `json:"first"`
	Last  string `json:"last"`
}
