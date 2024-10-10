package otp

type Storage interface {
	Create(args OTPArgs) (string, error)
	Read(id string) (string, error)
}
