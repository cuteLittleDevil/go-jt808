package command

import "encoding/xml"

type ConfirmType struct {
	XMLName xml.Name
	CmdType string `xml:"CmdType"`
}
