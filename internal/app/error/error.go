package error

import "errors"

var ErrLoginAlreadyTaken = errors.New("login already taken")
var ErrInvalidLoginOrPassword = errors.New("invalid login or password")
var ErrOrderLoaded = errors.New("order already loaded by user")
var ErrOrderInUse = errors.New("order alredy loaded by another user")
var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrNoWithdrawals = errors.New("no withdrawals")
var ErrNoOrders = errors.New("no orders")
