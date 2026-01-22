package dashboard

import (
	"os"
	"regexp"
	"strings"
)

func tmuxPrefix() string {
	prefix := os.Getenv("CRAIZY_TMUX_PREFIX")
	if prefix == "" {
		prefix = "craizy"
	}
	return prefix
}

func sessionName(agentName string) string {
	return tmuxPrefix() + "-" + slugify(agentName)
}

func sessionNameWithInstance(agentName, instance string) string {
	return tmuxPrefix() + "-" + slugify(agentName) + "-" + slugify(instance)
}

func slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "agent"
	}
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "agent"
	}
	return s
}
