package magiclink_test

import (
	"errors"
	"net/url"
	"testing"
	"time"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/mldtest"
)

func TestCreateParams_Valid(t *testing.T) {
	p := magiclink.CreateParams{}
	err := p.Valid()
	if !errors.Is(err, mld.ErrParams) {
		t.Errorf("expected error %s, got %s", mld.ErrParams, err)
	}

	p.Expires = time.Now().Add(mldtest.LinksExpireAfter)
	err = p.Valid()
	if !errors.Is(err, mld.ErrParams) {
		t.Errorf("expected error %s, got %s", mld.ErrParams, err)
	}
	p.Expires = time.Time{}

	p.RedirectURL = new(url.URL)
	err = p.Valid()
	if !errors.Is(err, mld.ErrParams) {
		t.Errorf("expected error %s, got %s", mld.ErrParams, err)
	}
	p.RedirectURL = nil

	p = magiclink.CreateParams{
		Expires:     time.Now().Add(mldtest.LinksExpireAfter),
		RedirectURL: new(url.URL),
	}
	err = p.Valid()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
}

func TestParams_Valid(t *testing.T) {
	p := magiclink.Config{}
	err := p.Valid()
	if !errors.Is(err, mld.ErrParams) {
		t.Errorf("expected error %s, got %s", mld.ErrParams, err)
	}

	p = magiclink.Config{
		ServiceURL: new(url.URL),
	}
	err = p.Valid()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
}
