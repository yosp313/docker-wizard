package utils

func OrderedSelectedIDs[T any](items []T, idOf func(T) string, selected map[string]bool) []string {
	ids := make([]string, 0, len(selected))
	for _, item := range items {
		id := idOf(item)
		if selected[id] {
			ids = append(ids, id)
		}
	}
	return ids
}
