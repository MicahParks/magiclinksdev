package magiclink_test

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/mldtest"
)

func TestCreateArgs_Valid(t *testing.T) {
	p := magiclink.CreateArgs{}
	err := p.Valid()
	if !errors.Is(err, magiclink.ErrArgs) {
		t.Errorf("expected error %s, got %s", magiclink.ErrArgs, err)
	}

	p.Expires = time.Now().Add(mldtest.LinksExpireAfter)
	err = p.Valid()
	if !errors.Is(err, magiclink.ErrArgs) {
		t.Errorf("expected error %s, got %s", magiclink.ErrArgs, err)
	}
	p.Expires = time.Time{}

	p.RedirectURL = new(url.URL)
	err = p.Valid()
	if !errors.Is(err, magiclink.ErrArgs) {
		t.Errorf("expected error %s, got %s", magiclink.ErrArgs, err)
	}
	p.RedirectURL = nil

	p = magiclink.CreateArgs{
		Expires:     time.Now().Add(mldtest.LinksExpireAfter),
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
