package pkg

import (
	"log"
	"testing"
)

func TestNewResolver(t *testing.T) {
	p := NewResolver()
	arg := []string{"www.sncf.com", "www.google.com"}
	res := p.Run(arg)
	log.Println(res)
}
