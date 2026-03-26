package wkutil

func ArrayContains(items []string, target string) bool {

	for _, element := range items {
		if target == element {
			return true

		}
	}
	return false

}

func ArrayEqual(items1 []string, items2 []string) bool {
	if len(items1) != len(items2) {
		return false
	}
	for i, element := range items1 {
		if element != items2[i] {
			return false
		}
	}
	return true
}
