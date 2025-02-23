package utils

func DoesUserHaveRequiredRole(roles []string) bool {
	for i := 0; i < len(roles); i++ {
		if roles[i] == "moder" || roles[i] == "admin" {
			return true
		}
	}
	return false
}
