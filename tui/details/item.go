package details

type Item interface {
	Details() View
}

type View struct {
	Title string
	Rows  []Row
}

type Row struct {
	Label string
	Value string
}
