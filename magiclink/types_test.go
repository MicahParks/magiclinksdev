package magiclink_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/MicahParks/magiclinksdev/magiclink"
)

func TestCreateArgs_Valid(t *testing.T) {
	p := magiclink.CreateArgs[any]{}
	err := p.Valid()
	if !errors.Is(err, magiclink.ErrArgs) {
		t.Errorf("expected error %s, got %s", magiclink.ErrArgs, err)
	}

	p = magiclink.CreateArgs[any]{
		RedirectURL: new(url.URL),
	}
	err = p.Valid()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
}

func TestArgs_Valid(t *testing.T) {
	p := magiclink.Config[any, any, any]{}
	err := p.Valid()
	if !errors.Is(err, magiclink.ErrArgs) {
		t.Errorf("expected error %s, got %s", magiclink.ErrArgs, err)
	}

	p = magiclink.Config[any, any, any]{
		ServiceURL: new(url.URL),
	}
	err = p.Valid()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
}
