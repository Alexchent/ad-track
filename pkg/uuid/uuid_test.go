package uuid

import (
	"fmt"
	"strings"
	"testing"
)

func TestBase(t *testing.T) {
	fmt.Println("=== 方法对比 ===")

	// V2: Base62
	fmt.Println("\nMakeNonceV2 (Base62):")
	for i := 0; i < 5; i++ {
		id := MakeNonce()
		fmt.Printf("  [%d] %s (len=%d, chars=%s)\n",
			i+1, id, len(id), getCharTypes(id))
	}

	// V4: 纯随机
	fmt.Println("\nMakeNonceV4 (纯随机):")
	for i := 0; i < 5; i++ {
		id := MakeNonceV2()
		fmt.Printf("  [%d] %s (len=%d, chars=%s)\n",
			i+1, id, len(id), getCharTypes(id))
	}
}

// 检查字符串包含的字符类型
func getCharTypes(s string) string {
	var hasDigit, hasLower, hasUpper bool
	for _, c := range s {
		if c >= '0' && c <= '9' {
			hasDigit = true
		} else if c >= 'a' && c <= 'z' {
			hasLower = true
		} else if c >= 'A' && c <= 'Z' {
			hasUpper = true
		}
	}

	var types []string
	if hasDigit {
		types = append(types, "数字")
	}
	if hasLower {
		types = append(types, "小写")
	}
	if hasUpper {
		types = append(types, "大写")
	}
	return strings.Join(types, "+")
}
