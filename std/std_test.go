package std

import "testing"

func Test_LimitsGen(t *testing.T) {
	file := `D:\App\go\src\gitee.com\suplxc\bak\data-helper\bin\std\sur1.csv`
	std, err := LimitsGen(file)
	t.Log(std)
	t.Log(err)
}
