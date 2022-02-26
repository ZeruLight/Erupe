package crypt

import (
	"fmt"
	"testing"
)

func TestCrypt(t *testing.T) {
	if ret := Crypt("testtest", "es"); ret != `esDRYJnY4VaGM` {
		t.Fatal(fmt.Sprintf(`result of Crypt is musmatch: %+v`, []byte(ret)))
	}
}

func ExampleCrypt() {
	fmt.Println(Crypt("testtest", "es"))
	// Output:
	// esDRYJnY4VaGM
}
