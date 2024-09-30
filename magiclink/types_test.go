package magiclink_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/MicahParks/magiclinksdev/magiclink"
)

func TestCreateArgs_Valid(t *testing.T) {
	p := magiclink.CreateArgs{}
	err := p.Valid()
	if !errors.Is(err, magiclink.ErrArgs) {
		t.Errorf("expected error %s, got %s", magiclink.ErrArgs, err)
	}

	p = magiclink.CreateArgs{
		RedirectURL: new(url.URL),
	}
	err = p.Valid()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
}

func TestArgs_Valid(t *testing.T) {
	p := magiclink.Config{}
	err := p.Valid()
	if !errors.Is(err, magiclink.ErrArgs) {
		t.Errorf("expected error %s, got %s", magiclink.ErrArgs, err)
	}

	p = magiclink.Config{
		ServiceURL: new(url.URL),
	}
	err = p.Valid()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
}
