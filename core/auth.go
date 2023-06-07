package core

func allowJoin(username string) (bool, string, error) {
	if cfg.Auth == "none" {
		return true, "", nil
	}

	if cfg.Auth == "whitelist" {
		for _, v := range cfg.Whitelist {
			if v == username {
				return true, "", nil
			}
		}
		return false, "You are not in the whitelist", nil
	}

	if cfg.Auth == "blacklist" {
		for _, v := range cfg.Blacklist {
			if v == username {
				return false, "You are in the blacklist", nil
			}
		}
		return true, "", nil
	}

	// should never reach here
	return false, "", nil
}
