package escaper_test

import (
	"BaiduPCS-Go/pcsutil/escaper"
	"fmt"
	"testing"
)

func TestEscape(t *testing.T) {
	fmt.Println(escaper.Escape(`asdf'asdfasd[]a[\[][sdf\[d]`, []rune{'[', '\''}))
}
