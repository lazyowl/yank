package ui

import "fmt"

// StdoutRecord stores the information that will be displayed to the user
type StdoutRecord struct {
	Names []string			// names of the file given by user
	FullHash string			// full hash
	PercentComplete int		// percent available (merged across all users)
	PercentLocal int		// percent I already have
	MaxComplete int			// maximum percentage of file I can have if I fetch it
	Users []string			// the users who contribute chunks
}

// StdoutTable is a list of StdoutRecords
type StdoutTable struct {
	Records []StdoutRecord
}

// NewTable creates a new StdoutTable
func NewTable() *StdoutTable {
	t := StdoutTable{}
	return &t
}

// Display prints out the contents of StdoutTable onto stdout
func (table *StdoutTable) Display() {
	fmt.Println("Names\t%c/l/m\tUsers\tFull hash\n")
	for _, r := range table.Records {
		fmt.Printf("%v\t%d/%d/%d\t%v\t%s\n", r.Names, r.PercentComplete, r.PercentLocal, r.MaxComplete, r.Users, r.FullHash)
	}
}
