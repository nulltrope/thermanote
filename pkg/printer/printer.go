package printer

type Printer interface {
	Print([]byte) error
}
