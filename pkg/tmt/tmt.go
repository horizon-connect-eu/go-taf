package tmt

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

/*
<typeDefinition>

	<type>A</type>
	<capacity>4</capacity>

</typeDefinition>
*/
type TypeDefinition struct {
	XMLName  xml.Name `xml:"typeDefinition"`
	TypeType string   `xml:"type"`
	Capacity int      `xml:"capacity"`
}

func ParseXmlFiles(xmlPath string, tmt map[string]int) error {
	err := filepath.Walk(xmlPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			xmlFile, err := os.Open(path)
			defer xmlFile.Close()
			if err == nil {
				fmt.Println("Successfully Opened " + path)

				byteValue, _ := io.ReadAll(xmlFile)
				var typeDef TypeDefinition
				xml.Unmarshal(byteValue, &typeDef)

				tmt[typeDef.TypeType] = typeDef.Capacity

			} else {
				fmt.Println(err)
			}
			return err
		}
		return nil
	})
	return err
}
