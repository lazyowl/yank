package ui

import "fmt"

type Stdout_record struct {
	Names []string				// names of the file given by user
	Full_hash string			// full hash
	Percent_complete int		// percent available (merged across all users)
	Percent_local int			// percent I already have
	Max_complete int			// maximum percentage of file I can have if I fetch it
	Users []string				// the users who contribute chunks
}

type Stdout_table struct {
	Records []Stdout_record
}

func New_table() *Stdout_table {
	t := Stdout_table{}
	return &t
}

func (table *Stdout_table) Display() {
	fmt.Println("Names\n%c/l/m\nUsers\nFull hash\n")
	for _, r := range table.Records {
		fmt.Printf("%v\n%d/%d/%d\n%v\n%s\n\n", r.Names, r.Percent_complete, r.Percent_local, r.Max_complete, r.Users, r.Full_hash)
	}
}
