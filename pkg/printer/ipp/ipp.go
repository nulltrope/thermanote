package ipp

import (
	"bytes"

	"github.com/phin1x/go-ipp"
)

type IPPPrinter struct {
	cli *ipp.CUPSClient
}

func NewIPPPrinter() *IPPPrinter {
	return &IPPPrinter{
		cli: ipp.NewCUPSClient("192.168.0.75", 631, "jkelly", "foobar", true),
	}
}

func (p *IPPPrinter) Print(content []byte) error {
	buf := bytes.NewBuffer(content)

	jobAttr := map[string]interface{}{}
	jobAttr[ipp.AttributeJobName] = "foo1234"
	jobAttr[ipp.AttributeOrientationRequested] = 3 // Portrait

	_, err := p.cli.PrintDocuments([]ipp.Document{
		{
			Document: buf,
			Name:     "foo1234",
			Size:     int(buf.Len()),
			MimeType: ipp.MimeTypeOctetStream,
		},
	}, "Zebra_Technologies_ZTC_LP2844-Z-200dpi", jobAttr)

	return err
}
