package mpq

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Pool pre-loads multiple MPQ archive headers, allowing fast, concurrent access of archive files
type Pool struct {
	// Archive data
	fmap map[string]*archiveEntry
}

type archiveEntry struct {
	name       string
	header     *Header
	hashTable  []*HashEntry
	blockTable []*BlockEntry
}

func getArchiveName(name string) string {
	s := strings.Split(name, "/")
	return s[len(s)-1]
}

func (p *Pool) addArchive(name string) error {
	m, err := Open(name)
	if err != nil {
		return err
	}

	fmt.Println("[MPQ Pool] Opened", name)

	ae := new(archiveEntry)
	ae.name = name
	ae.header = m.Header
	ae.hashTable = m.HashTable
	ae.blockTable = m.BlockTable

	lf := m.ListFiles()

	for _, fv := range lf {
		mappedFile := p.fmap[fv]
		if mappedFile == nil {
			p.fmap[fv] = ae // map filepath string to MPQ data pointer
		}
	}

	return nil
}

// OpenPool opens a Pool using a slice of MPQ file paths
func OpenPool(names []string) (*Pool, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("mpq: cannot open Pool without at least one archive")
	}

	p := &Pool{}
	p.fmap = make(map[string]*archiveEntry)

	// Add patch archives first.
	var patch, todo []string
	for _, v := range names {
		if strings.Contains(getArchiveName(v), "patch") {
			patch = append(patch, v)
		} else {
			todo = append(todo, v)
		}
	}

	sort.Strings(patch)
	for _, v := range patch {
		err := p.addArchive(v)
		if err != nil {
			return nil, err
		}
	}

	// Add other archives later.
	for _, v := range todo {
		err := p.addArchive(v)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *Pool) OpenFile(name string) (*File, error) {
	ae := p.fmap[name]
	if ae == nil {
		return nil, fmt.Errorf("File not found")
	}

	m := new(MPQ)
	m.Path = ae.name
	var err error
	m.File, err = os.Open(ae.name)
	if err != nil {
		return nil, err
	}

	m.Header = ae.header
	m.BlockTable = ae.blockTable
	m.HashTable = ae.hashTable

	file, err := m.OpenFile(name)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (p *Pool) ListFiles() []string {
	str := make([]string, len(p.fmap))
	i := 0
	for k := range p.fmap {
		str[i] = k
		i++
	}

	sort.Strings(str)

	return str
}
