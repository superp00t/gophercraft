package mpq

import (
	"path/filepath"
)

var Patterns = []string{
	// 5875,
	"backup.MPQ",
	"base.MPQ",
	"dbc.MPQ",
	"fonts.MPQ",
	"interface.MPQ",
	"misc.MPQ",
	"model.MPQ",
	"patch.MPQ",
	"patch-2.MPQ",
	"sound.MPQ",
	"speech.MPQ",
	"terrain.MPQ",
	"texture.MPQ",
	"wmo.MPQ",

	// 12340
	"common.MPQ",
	"common-2.MPQ",
	"expansion.MPQ",
	"lichking.MPQ",
	"*/locale-*.MPQ",
	"*/speech-*.MPQ",
	"*/expansion-locale-*.MPQ",
	"*/lichking-locale-*.MPQ",
	"*/expansion-speech-*.MPQ",
	"*/lichking-speech-*.MPQ",
	"*/patch-*.MPQ",
	"patch.MPQ",
	"patch-*.MPQ",
}

func GetFiles(basepath string) ([]string, error) {
	var mpqFiles []string

	for _, v := range Patterns {
		fp := filepath.Join(basepath, v)
		m, err := filepath.Glob(fp)
		if err != nil {
			return nil, err
		}

	matchLoop:
		for _, match := range m {
			for _, file := range mpqFiles {
				if match == file {
					continue matchLoop
				}
			}

			mpqFiles = append(mpqFiles, match)
			break
		}
	}

	return mpqFiles, nil
}
