package errs

import "errors"

var ErrUnauthorized = errors.New("未授权")
var ErrSessionKeyNotFound = errors.New("session 中没找到对应的 key")
