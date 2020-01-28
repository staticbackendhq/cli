package cmd

type byID []map[string]interface{}

func (x byID) Len() int {
	return len(x)
}

func (x byID) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x byID) Less(i, j int) bool {
	a, ok := x[i]["id"].(int)
	if !ok {
		return false
	}

	b, ok := x[j]["id"].(int)
	if !ok {
		return false
	}

	return a < b
}

type byIDDesc []map[string]interface{}

func (x byIDDesc) Len() int {
	return len(x)
}

func (x byIDDesc) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x byIDDesc) Less(i, j int) bool {
	a, ok := x[i]["id"].(int)
	if !ok {
		return false
	}

	b, ok := x[j]["id"].(int)
	if !ok {
		return false
	}

	return a > b
}
