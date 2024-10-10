package otp

type CreateParams struct {
}

type Storage interface {
	Create(args OTPParams) (string, error)
	Read(id string) (string, error)
}
