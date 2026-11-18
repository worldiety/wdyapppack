// Package plist reads and writes Apple property lists in their XML form.
//
// Values map to Go as follows: dict ↔ map[string]any, array ↔ []any,
// string ↔ string, integer ↔ int64, real ↔ float64, true/false ↔ bool,
// date ↔ time.Time, data ↔ []byte. Encode also accepts the other integer
// types, []string and map[string]string for convenience.
//
// Keys are written in sorted order, so the same values always produce the same
// bytes. The binary form (bplist00) is recognised and refused with ErrBinary:
// the files this package reads are written by people, and a person who has a
// binary one can convert it with plutil.
package plist

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

// ErrBinary is returned when the data is a binary property list.
var ErrBinary = errors.New("binary property list; convert it with: plutil -convert xml1 <file>")

const header = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
`

// Encode renders v as an XML property list.
func Encode(v any) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(header)
	if err := encode(&b, v, 0); err != nil {
		return nil, err
	}
	b.WriteString("</plist>\n")
	return b.Bytes(), nil
}

func encode(b *bytes.Buffer, v any, depth int) error {
	ind := strings.Repeat("\t", depth)
	switch x := v.(type) {
	case map[string]any:
		if len(x) == 0 {
			b.WriteString(ind + "<dict/>\n")
			return nil
		}
		b.WriteString(ind + "<dict>\n")
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			b.WriteString(ind + "\t<key>" + escape(k) + "</key>\n")
			if err := encode(b, x[k], depth+1); err != nil {
				return fmt.Errorf("%s: %w", k, err)
			}
		}
		b.WriteString(ind + "</dict>\n")
	case map[string]string:
		m := make(map[string]any, len(x))
		for k, s := range x {
			m[k] = s
		}
		return encode(b, m, depth)
	case []any:
		if len(x) == 0 {
			b.WriteString(ind + "<array/>\n")
			return nil
		}
		b.WriteString(ind + "<array>\n")
		for i, e := range x {
			if err := encode(b, e, depth+1); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
		b.WriteString(ind + "</array>\n")
	case []string:
		a := make([]any, len(x))
		for i, s := range x {
			a[i] = s
		}
		return encode(b, a, depth)
	case string:
		b.WriteString(ind + "<string>" + escape(x) + "</string>\n")
	case bool:
		if x {
			b.WriteString(ind + "<true/>\n")
		} else {
			b.WriteString(ind + "<false/>\n")
		}
	case int:
		return encode(b, int64(x), depth)
	case int32:
		return encode(b, int64(x), depth)
	case uint32:
		return encode(b, int64(x), depth)
	case int64:
		b.WriteString(ind + "<integer>" + strconv.FormatInt(x, 10) + "</integer>\n")
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return fmt.Errorf("a property list cannot hold %v", x)
		}
		b.WriteString(ind + "<real>" + strconv.FormatFloat(x, 'g', -1, 64) + "</real>\n")
	case time.Time:
		b.WriteString(ind + "<date>" + x.UTC().Format("2006-01-02T15:04:05Z") + "</date>\n")
	case []byte:
		b.WriteString(ind + "<data>" + base64.StdEncoding.EncodeToString(x) + "</data>\n")
	default:
		return fmt.Errorf("a property list cannot hold a value of type %T", v)
	}
	return nil
}

func escape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

// Decode parses an XML property list.
func Decode(data []byte) (any, error) {
	if bytes.HasPrefix(data, []byte("bplist")) {
		return nil, ErrBinary
	}
	d := xml.NewDecoder(bytes.NewReader(data))
	d.Strict = true
	for {
		tok, err := d.Token()
		if err != nil {
			if err == io.EOF {
				return nil, errors.New("no <plist> element")
			}
			return nil, positioned(d, err)
		}
		if se, ok := tok.(xml.StartElement); ok {
			if se.Name.Local != "plist" {
				return nil, fmt.Errorf("the root element is <%s>, not <plist>", se.Name.Local)
			}
			v, end, err := value(d)
			if err != nil {
				return nil, positioned(d, err)
			}
			if end {
				return nil, errors.New("the <plist> element is empty")
			}
			return v, nil
		}
	}
}

func positioned(d *xml.Decoder, err error) error {
	line, _ := d.InputPos()
	return fmt.Errorf("line %d: %w", line, err)
}

// value reads the next value. end reports that the enclosing element closed
// instead.
func value(d *xml.Decoder) (any, bool, error) {
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, false, err
		}
		switch t := tok.(type) {
		case xml.EndElement:
			return nil, true, nil
		case xml.StartElement:
			v, err := element(d, t)
			return v, false, err
		}
	}
}

func element(d *xml.Decoder, se xml.StartElement) (any, error) {
	switch se.Name.Local {
	case "dict":
		m := map[string]any{}
		for {
			tok, err := d.Token()
			if err != nil {
				return nil, err
			}
			switch t := tok.(type) {
			case xml.EndElement:
				return m, nil
			case xml.StartElement:
				if t.Name.Local != "key" {
					return nil, fmt.Errorf("expected <key> in <dict>, found <%s>", t.Name.Local)
				}
				k, err := text(d)
				if err != nil {
					return nil, err
				}
				v, end, err := value(d)
				if err != nil {
					return nil, err
				}
				if end {
					return nil, fmt.Errorf("key %q has no value", k)
				}
				m[k] = v
			}
		}
	case "array":
		a := []any{}
		for {
			v, end, err := value(d)
			if err != nil {
				return nil, err
			}
			if end {
				return a, nil
			}
			a = append(a, v)
		}
	case "string":
		return text(d)
	case "integer":
		s, err := text(d)
		if err != nil {
			return nil, err
		}
		return strconv.ParseInt(strings.TrimSpace(s), 0, 64)
	case "real":
		s, err := text(d)
		if err != nil {
			return nil, err
		}
		return strconv.ParseFloat(strings.TrimSpace(s), 64)
	case "true", "false":
		if err := d.Skip(); err != nil {
			return nil, err
		}
		return se.Name.Local == "true", nil
	case "date":
		s, err := text(d)
		if err != nil {
			return nil, err
		}
		return time.Parse(time.RFC3339, strings.TrimSpace(s))
	case "data":
		s, err := text(d)
		if err != nil {
			return nil, err
		}
		return base64.StdEncoding.DecodeString(strings.Join(strings.Fields(s), ""))
	}
	return nil, fmt.Errorf("unknown element <%s>", se.Name.Local)
}

// text reads character data up to the end of the current element.
func text(d *xml.Decoder) (string, error) {
	var b strings.Builder
	for {
		tok, err := d.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.CharData:
			b.Write(t)
		case xml.EndElement:
			return b.String(), nil
		case xml.StartElement:
			return "", fmt.Errorf("unexpected <%s> inside a text value", t.Name.Local)
		}
	}
}
