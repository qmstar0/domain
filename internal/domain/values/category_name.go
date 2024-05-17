package values

import (
	"errors"
	"regexp"
	"strings"
)

var CategoryRegexpCheck, _ = regexp.Compile(`^[\p{L}\p{N}]+$`)

type CategoryName string

func NewCategoryName(s string) (CategoryName, error) {
	s = strings.TrimSpace(s)
	if !CategoryRegexpCheck.MatchString(s) {
		return "", errors.New("分类名格式错误")
	}
	return CategoryName(s), nil
}

func (n CategoryName) String() string {
	return string(n)
}