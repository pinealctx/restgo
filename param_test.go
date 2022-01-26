package restgo

import "testing"

type A struct {
	List []string `query:"list"`
}

func Test_ObjectParam(t *testing.T) {
	var params = ObjectParams(&A{List: []string{"a", "b"}})
	t.Log(params)
}
