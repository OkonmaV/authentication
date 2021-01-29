package main

type Authentication interface {
	Valid(login, pass string) error
}
