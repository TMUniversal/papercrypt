package codematrix

import "errors"

var (
	errSAHeaderTooShort   = errors.New("codematrix: SA header too short")
	errInvalidSAMode      = errors.New("codematrix: invalid SA mode indicator")
	errDataHeaderShort    = errors.New("codematrix: data header too short")
	errInvalidByteMode    = errors.New("codematrix: invalid byte mode indicator")
	errPayloadTooLarge    = errors.New("codematrix: payload too large for QR version 40-H")
	errNoImages           = errors.New("codematrix: no images provided")
	errTooManyImages      = errors.New("codematrix: too many images")
	errDataBitsMismatch   = errors.New("codematrix: data bits size mismatch")
	errInterleaveMismatch = errors.New("codematrix: interleaved size mismatch")
	errNoValidMask        = errors.New("codematrix: no valid mask pattern found")
)
