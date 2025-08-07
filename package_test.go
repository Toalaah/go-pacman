package pacman_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	pacman "github.com/toalaah/go-pacman"
)

func TestPackageTextMarshal(t *testing.T) {
	assert := assert.New(t)
	pkg := makeStubPackage()
	expected, err := os.ReadFile("./fixtures/xz.desc")
	assert.Nil(err)

	got, err := pkg.MarshalText()
	assert.Nil(err)

	assert.Equal(expected, got, "Expected marshaled package desciption to equal fixture")
}

func TestPackageTextUnmarshal(t *testing.T) {
	assert := assert.New(t)

	expected := makeStubPackage()
	got := pacman.Package{}

	b, err := os.ReadFile("./fixtures/xz.desc")
	assert.Nil(err)

	err = got.UnmarshalText(b)
	assert.Nil(err)

	assert.Equal(expected, got, "Expected unmarshaled package desciption to equal fixture")
}

func makeStubPackage() pacman.Package {
	return pacman.Package{
		FileName:    "xz-5.8.1-1-x86_64.pkg.tar.zst",
		Name:        "xz",
		Base:        "xz",
		Version:     "5.8.1-1",
		Description: "Library and command line tools for XZ and LZMA compressed files",
		CSize:       831572,
		ISize:       3060622,
		Sha256Sum: [32]byte{
			0xae, 0xec, 0xc6, 0x31, 0x5b, 0x7b, 0x6d, 0x6a, 0xf8, 0xd4, 0x3a, 0x37, 0x5b, 0x1e, 0x27, 0x95, 0xe3, 0x13, 0x56, 0x3b, 0xfd, 0xe8, 0xa5, 0xdf, 0x58, 0x66, 0x95, 0x2b, 0x08, 0x7e, 0xb6, 0xad,
		},
		Arch:         pacman.Amd64,
		Licenses:     []pacman.License{pacman.GPL, pacman.LGPL, pacman.Custom},
		URL:          "https://tukaani.org/xz/",
		BuildDate:    time.Unix(1743698592, 0),
		PGPSignature: "iQIzBAABCgAdFiEE4kC1fixGMLp2ji8m/BtUfI2BcsgFAmfuu34ACgkQ/BtUfI2BcsgpcA/+IrA2GDgQICXAGBapp3YPLgo8Gw7b9kmsi9j/iY27tV7IuioYCEpHnEt7fSMggSh8svg9wPKHGaElJdGjcT3lu/p/0xQXryRuFdf9jX6NdEnODYLUIOITIVZNzcQOUtddr4y5P88gd7aXnY2OBbmhSvbMVCkwzkpwSdYkWj6gp7Gi/4kBiEKgToFYkrC2xd0lBxEDbDurYAGwW90fdVCOW16Mlu4ysI49y6sx8YLpT4QmA2Yy3DrIE824dONdEoYExK6gzYVyhLu7F1gpv6Nwy1WHCAj5jo3+cmMlpWTfxzyuDMjb3Bg9N5ZZjU8L5SgNPFXo3g9uwEZvbB/tufHigc5Ss4X4ctwfIVdcQgmbcvOwzMNlwfte6upXSfSqryijy2f16zmhyxJdV55E24NLxmsglUEMRBBriv/gOl7pV61lpOa6pcjC1xnwAob6bkJFSP2KfgDtiQAGOhV+wwRy0bUmsDC+x9t6pOazmgYvCQKrdAefRi/QWJTVwh1FFQLlz4TP/K7jsT7fMowjr/5Gjeg/s0TKY8vtsapwdlTbf3zJFwF9m1/MQY4LungPBmXnFLyI+TH3kzvONMg/EaAe4z9R2Brba8H3TPWfXbST+KjD24ZBYpd2RRoB6f1hjxEUsqG5asynO6iOEeQf72nzWcMFhkoXUZa/+1nmbj9s71k=",
		Packager: pacman.Packager{
			Name:  "Levente Polyak",
			Email: "anthraxx@archlinux.org",
		},
		Provides: []string{
			"liblzma.so=5-64",
		},
		Depends: []string{
			"sh",
		},
		MakeDepends: []string{
			"git",
			"po4a",
			"doxygen",
		},
	}
}
