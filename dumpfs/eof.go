package dumpfs

type OnEOF func(filename string) error

func KeepFile(_ string) error {
	return nil
}
