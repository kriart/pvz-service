package reception

func CanOpenNew(last *Reception) bool {
	if last == nil {
		return true
	}
	return last.Status != StatusInProgress
}
