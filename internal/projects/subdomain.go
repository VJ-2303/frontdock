package projects

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
)

var subdomainRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,61}[a-z0-9])?$`)

var reserved = map[string]bool{
	"www": true, "api": true, "app": true, "admin": true, "edge": true,
	"mail": true, "smtp": true, "ftp": true, "cdn": true, "static": true,
	"assets": true, "docs": true, "blog": true, "status": true,
	"dashboard": true,
	"login":     true, "auth": true, "account": true, "billing": true,
	"support":   true,
	"frontdock": true, "internal": true, "test": true, "staging": true,
	"dev": true,
}

func ValidateSubdomains(s string) error {
	if len(s) < 3 || len(s) > 63 {
		return fmt.Errorf("subdomain must be 3-63 characters")
	}
	if !subdomainRegex.MatchString(s) {
		return fmt.Errorf("subdomain may contain only lowercase letters, digits adn hypens, and cannot start or end with a hyphen")
	}
	if reserved[s] {
		return fmt.Errorf("that subdomain is reserved")
	}
	return nil
}

var adjectives = []string{"swift", "calm", "bold", "brave", "keen", "warm", "cool", "bright", "quiet", "clever"}
var nouns = []string{"otter", "falcon", "cedar", "harbor", "meadow", "canyon", "lantern", "comet", "willow", "ember"}

func GenerateSubdomain() string {
	pick := func(list []string) string {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(list))))
		return list[n.Int64()]
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	return fmt.Sprintf("%s-%s-%d", pick(adjectives), pick(nouns),
		n.Int64()+1000)
}
