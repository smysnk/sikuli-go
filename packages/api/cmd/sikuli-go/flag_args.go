package main

import "strings"

const (
	defaultGRPCListenAddr  = ":50051"
	defaultAdminListenAddr = ":8080"
)

func normalizeServerFlagArgs(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	out := make([]string, 0, len(args)+2)
	for idx := 0; idx < len(args); idx++ {
		arg := args[idx]
		switch arg {
		case "-listen", "--listen":
			out = append(out, arg)
			if nextFlagValueProvided(args, idx) {
				continue
			}
			out = append(out, defaultGRPCListenAddr)
		case "-admin-listen", "--admin-listen":
			out = append(out, arg)
			if nextFlagValueProvided(args, idx) {
				continue
			}
			out = append(out, defaultAdminListenAddr)
		default:
			out = append(out, arg)
		}
	}
	return out
}

func nextFlagValueProvided(args []string, idx int) bool {
	if idx+1 >= len(args) {
		return false
	}
	next := strings.TrimSpace(args[idx+1])
	if next == "" {
		return false
	}
	if strings.HasPrefix(next, "-") {
		return false
	}
	return true
}
