package pacman

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const tagName = "pacman"
const headerFormat = "%%%s%%\n"

type Package struct {
	FileName     string            `pacman:"header=FILENAME,omitempty"`
	Name         string            `pacman:"header=NAME,omitempty"`
	Base         string            `pacman:"header=BASE,omitempty"`
	Version      string            `pacman:"header=VERSION,omitempty"`
	Description  string            `pacman:"header=DESC,omitempty"`
	CSize        uint64            `pacman:"header=CSIZE,omitempty"`
	ISize        uint64            `pacman:"header=ISIZE,omitempty"`
	Sha256Sum    [sha256.Size]byte `pacman:"header=SHA256SUM,omitempty"`
	PGPSignature string            `pacman:"header=PGPSIG,omitempty"`
	URL          string            `pacman:"header=URL,omitempty"`
	Licenses     []License         `pacman:"header=LICENSE,omitempty"`
	Arch         Architecture      `pacman:"header=ARCH,omitempty"`
	BuildDate    time.Time         `pacman:"header=BUILDDATE,omitempty"`
	Packager     Packager          `pacman:"header=PACKAGER,omitempty"`
	Provides     []string          `pacman:"header=PROVIDES,omitempty"`
	Depends      []string          `pacman:"header=DEPENDS,omitempty"`
	MakeDepends  []string          `pacman:"header=MAKEDEPENDS,omitempty"`
	OptDepends   []OptDependency   `pacman:"header=OPTDEPENDS,omitempty"`
	CheckDepends []string          `pacman:"header=CHECKDEPENDS,omitempty"`
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (p *Package) UnmarshalText(text []byte) error {
	sc := bufio.NewScanner(bytes.NewReader(text))
	sc.Split(splitSection)
	for sc.Scan() {
		s := sc.Text()
		lines := strings.Split(s, "\n")
		if err := p.parseSection(lines); err != nil {
			return err
		}
	}
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface
func (p *Package) MarshalText() ([]byte, error) {
	b := bytes.Buffer{}
	// We use reflection here to save on some of the redundant serialization that we would otherwise have to perform. Though, whether this is worth the (likely) performance penatly is somewhat dubious.
	t := reflect.TypeOf(p).Elem()
	v := reflect.Indirect(reflect.ValueOf(p))
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		st := getStructTag(field)
		writeSection(&b, v.Field(i).Interface(), st)
	}
	return b.Bytes(), nil
}

type Architecture string

const (
	Amd64 Architecture = "x86_64"
)

func (a Architecture) String() string {
	return string(a)
}

type License string

func (l License) String() string {
	return string(l)
}

const (
	GPL    License = "GPL"
	LGPL   License = "LGPL"
	Custom License = "custom"
)

type OptDependency struct {
	Package string
	Reason  string
}

// String implements the fmt.Stringer interface
func (d OptDependency) String() string {
	s := d.Package
	if d.Reason != "" {
		s += ": " + d.Reason
	}
	return s
}

type Packager struct {
	Email string
	Name  string
}

// String implements the fmt.Stringer interface
func (p Packager) String() string {
	s := p.Name
	if p.Email != "" {
		s += " <" + p.Email + ">"
	}
	return s
}

type structTag struct {
	Header    string
	OmitEmpty bool
}

func getStructTag(field reflect.StructField) structTag {
	tag := field.Tag.Get(tagName)
	st := structTag{}
	for opt := range strings.SplitSeq(tag, ",") {
		opt := strings.TrimSpace(opt)
		sp := strings.Split(opt, "=")
		// Option may not have an assignment, even so we still split on "=" as it gives us the identifier either way.
		switch sp[0] {
		case "omitempty":
			st.OmitEmpty = true
		case "header":
			if len(sp) != 2 {
				panic("expected header tag to have assiged value")
			}
			st.Header = sp[1]
		default:
			panic("unknown tag: " + opt)
		}
	}
	return st
}

func writeSection(b *bytes.Buffer, v any, opts structTag) {
	if opts.OmitEmpty && reflect.Indirect(reflect.ValueOf(v)).IsZero() {
		return
	}
	// Write header
	fmt.Fprintf(b, headerFormat, opts.Header)
	// Write contents
	switch opts.Header {
	case "FILENAME", "NAME", "BASE", "VERSION", "DESC", "PGPSIG", "URL", "ARCH", "PACKAGER":
		fmt.Fprintf(b, "%s", v)
	case "CSIZE", "ISIZE":
		b.WriteString(strconv.FormatUint(v.(uint64), 10))
	case "BUILDDATE":
		v := v.(time.Time)
		b.WriteString(strconv.FormatUint(uint64(v.Unix()), 10))
	case "SHA256SUM":
		v := v.([sha256.Size]byte)
		b.WriteString(hex.EncodeToString(v[:]))
	case "LICENSE":
		v := v.([]License)
		writeArray(b, v)
	case "PROVIDES", "DEPENDS", "MAKEDEPENDS", "CHECKDEPENDS":
		v := v.([]string)
		b.WriteString(strings.Join(v, "\n"))
	case "OPTDEPENDS":
		v := v.([]OptDependency)
		writeArray(b, v)
	// Yeah, we could just not care here and write whatever comes out of fmt.Sprintf("%+v", v), but since this is internal only we can expect to never run into this case. Hence, if we ever do, we are in big trouble anyways and/or forgot to account for a newly added header.
	default:
		panic("unknown header, do not know how to format: " + opts.Header)
	}
	// Write epilogue
	b.WriteString("\n\n")
}

func splitSection(data []byte, atEOF bool) (advance int, token []byte, err error) {
	search := []byte("\n\n%")
	searchLen := len(search)
	l := len(data)
	if atEOF && l == 0 {
		return 0, nil, nil
	}
	if string(data[0]) != "%" {
		return 0, nil, fmt.Errorf("expected delimiter")
	}
	// New section start.
	// Read until next header delimiter
	i := bytes.Index(data[1:], search)
	if i > 0 {
		return i + searchLen, data[0 : i+1], nil
	}
	// No further section found, we are either in last section or at EOF
	if !atEOF {
		return l, data[:len(data)-searchLen+1], nil
	}
	// EOF
	return l, data, nil

}

func (p *Package) parseSection(section []string) error {
	if len(section) < 2 {
		return fmt.Errorf("unexpected section length: %s", section)
	}
	header := strings.TrimLeft(strings.TrimRight(section[0], "%"), "%")
	data := strings.Join(section[1:], "\n")
	var err error
	switch header {
	// String types
	case "FILENAME":
		p.FileName = data
	case "NAME":
		p.Name = data
	case "BASE":
		p.Base = data
	case "VERSION":
		p.Version = data
	case "DESC":
		p.Description = data
	case "PGPSIG":
		p.PGPSignature = data
	case "URL":
		p.URL = data
	case "ARCH":
		// TODO: validate architecture.
		p.Arch = Architecture(data)

	// Int types
	case "CSIZE":
		p.CSize, err = strconv.ParseUint(data, 10, 64)
		if err != nil {
			return err
		}
	case "ISIZE":
		p.ISize, err = strconv.ParseUint(data, 10, 64)
		if err != nil {
			return err
		}

	// List types
	case "LICENSE":
		for _, e := range section[1:] {
			// TODO: validate licenses
			p.Licenses = append(p.Licenses, License(e))
		}
	case "PROVIDES":
		p.Provides = append(p.Provides, section[1:]...)
	case "DEPENDS":
		p.Depends = append(p.Depends, section[1:]...)
	case "MAKEDEPENDS":
		p.MakeDepends = append(p.MakeDepends, section[1:]...)
	case "OPTDEPENDS":
		for _, e := range section[1:] {
			optDep := strings.Split(e, ": ")
			if len(optDep) != 2 {
				return fmt.Errorf("unexpected structure for opt dependency: %s", e)
			}
			p.OptDepends = append(p.OptDepends, OptDependency{
				Package: optDep[0],
				Reason:  optDep[1],
			})
		}
	case "CHECKDEPENDS":
		p.CheckDepends = append(p.CheckDepends, section[1:]...)
	case "SHA256SUM":
		sum, err := hex.DecodeString(data)
		if err != nil {
			return err
		}
		copy(p.Sha256Sum[:], sum[:sha256.Size])

	// Misc types
	case "BUILDDATE":
		var t int64
		t, err = strconv.ParseInt(data, 10, 64)
		if err != nil {
			return err
		}
		p.BuildDate = time.Unix(t, 0)
	case "PACKAGER":
		d := strings.Split(data, " ")
		p.Packager.Name = strings.Join(d[:len(d)-1], " ")
		p.Packager.Email = strings.TrimRight(strings.TrimLeft(d[len(d)-1], "<"), ">")
	default:
		return fmt.Errorf("unknown header: %s", header)
	}

	return nil
}

func writeArray[T fmt.Stringer](b *bytes.Buffer, v []T) {
	l := len(v)
	for i, e := range v {
		fmt.Fprintf(b, "%s", e)
		if i < l-1 {
			b.WriteString("\n")
		}
	}
}
