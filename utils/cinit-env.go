package utils

import (
	"os"
	"strings"
)

// KcadmEnv for reading Kcadm env into a map
type KcadmEnv map[string]string

// GetKcadmEnv will read environ and create a map of k:v from envs
// that have a KCADM_ prefix. The prefix is removed
func GetKcadmEnv() map[string]string {
	var key string
	env := make(KcadmEnv)
	osEnviron := os.Environ()
	cinitPrefix := "KCADM_"
	for _, b := range osEnviron {
		if strings.HasPrefix(b, cinitPrefix) {
			pair := strings.SplitN(b, "=", 2)
			key = strings.TrimPrefix(pair[0], cinitPrefix)
			key = strings.ToLower(key)
			key = strings.Replace(key, "_", ".", -1)
			env[key] = pair[1]
		}
	}

	return env
}
