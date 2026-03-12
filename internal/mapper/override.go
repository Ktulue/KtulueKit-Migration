package mapper

// ApplyDestOverride rewrites resolvedTarget according to destRoot.
//
// Guard order:
//  1. Empty destRoot → return resolvedTarget unchanged.
//  2. resolvedTarget does not match [A-Za-z]:\ → return unchanged.
//  3. len(destRoot)==3 (e.g. "D:\") → drive-letter swap.
//  4. Otherwise → strip X:\ prefix and prepend destRoot.
func ApplyDestOverride(resolvedTarget, destRoot string) string {
	if destRoot == "" {
		return resolvedTarget
	}
	// Guard 2: require X:\ prefix on target
	if len(resolvedTarget) < 3 || resolvedTarget[1] != ':' || resolvedTarget[2] != '\\' {
		return resolvedTarget
	}
	if len(destRoot) == 3 {
		// Drive swap: replace first character (drive letter)
		return string(destRoot[0]) + resolvedTarget[1:]
	}
	// Prefix substitution: strip "X:\" (first 3 chars) and prepend destRoot.
	// Ensure destRoot ends with \ so we don't join without a separator.
	if len(destRoot) > 0 && destRoot[len(destRoot)-1] != '\\' {
		destRoot += "\\"
	}
	return destRoot + resolvedTarget[3:]
}
