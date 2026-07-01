package service

func calcStatus(score int) string {
	switch {
	case score >= 95:
		return "GREEN"
	case score >= 90:
		return "BLUE"
	case score >= 80:
		return "YELLOW"
	case score >= 70:
		return "ORANGE"
	case score >= 60:
		return "RED"
	default:
		return "DEEP_REPAIR"
	}
}
