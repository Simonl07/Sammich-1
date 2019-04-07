package data

type Identity struct {
	Name    string
	Age     int32
	Address string
	Email   string
	Phone   string
}

type Merits struct {
	Skills     []string
	Education  []string
	Experience []string
}

func NewIdentity(name string, age int32, address string, email string, phone string) *Identity {
	return &Identity{Name: name, Age: age, Address: address, Email: email, Phone: phone}
}

func NewMerits(skills []string, education []string, experience []string) *Merits {
	return &Merits{Skills: skills, Education: education, Experience: experience}
}
