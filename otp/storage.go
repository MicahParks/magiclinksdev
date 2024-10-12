package otp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrOTPInvalid = errors.New("OTP invalid")

type CreateParams struct {
	CharSetAlphaLower bool
	CharSetAlphaUpper bool
	CharSetNumeric    bool
	Expires           time.Time
	Length            int64
}

type CreateResult struct {
	CreateParams CreateParams
	ID           string
	OTP          string
}

type Storage interface {
	OTPCreate(ctx context.Context, params CreateParams) (CreateResult, error)
	OTPValidate(ctx context.Context, id, otp string) error
}

type memoryOTP struct {
	mux   sync.Mutex
	store map[string]CreateResult
}

func NewMemoryStorage() Storage {
	return &memoryOTP{
		store: make(map[string]CreateResult),
	}
}
func (m *memoryOTP) OTPCreate(_ context.Context, params CreateParams) (CreateResult, error) {
	o, err := generateOTP(params)
	if err != nil {
		return CreateResult{}, fmt.Errorf("failed to generate OTP: %w", err)
	}
	id := uuid.New().String()
	result := CreateResult{
		CreateParams: params,
		ID:           id,
		OTP:          o,
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	m.store[id] = result
	return result, nil
}
func (m *memoryOTP) OTPValidate(_ context.Context, id, otp string) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	o, ok := m.store[id]
	if !ok {
		return fmt.Errorf("OTP not found: %w", ErrOTPInvalid)
	}
	if o.CreateParams.Expires.Before(time.Now()) {
		return fmt.Errorf("OTP expired: %w", ErrOTPInvalid)
	}
	if o.OTP != otp {
		return fmt.Errorf("OTP incorrect: %w", ErrOTPInvalid)
	}
	delete(m.store, id)
	return nil
}
