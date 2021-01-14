package mpq

import "fmt"

type FileCorruptionError struct {
	Filename        string
	CorruptionError error
}

func (fce FileCorruptionError) Error() string {
	return fmt.Sprintf("mpq: file %s was found to be corrupt: \"%s\"", fce.Filename, fce.CorruptionError)
}

type FileWasDeletedError string

func (fwde FileWasDeletedError) Error() string {
	return "mpq: file was deleted: " + string(fwde)
}
