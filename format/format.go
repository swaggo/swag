package format

import (
	"fmt"
	"log"
	"os"

	"github.com/swaggo/swag"
)

type Fmt struct {
}

func New() *Fmt {
	return &Fmt{}
}

type Config struct {
	// SearchDir the swag would be parse
	SearchDir string

	// excludes dirs and files in SearchDir,comma separated
	Excludes string

	MainFile string
}

func (f *Fmt) Build(config *Config) error {
	if _, err := os.Stat(config.SearchDir); os.IsNotExist(err) {
		return fmt.Errorf("dir: %s is not exist", config.SearchDir)
	}

	log.Println("Formating code.... ")
	formater := swag.NewFormater()
	if err := formater.FormatAPI(config.SearchDir, config.Excludes, config.MainFile); err != nil {
		return err
	}
	return nil
}
