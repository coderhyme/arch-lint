package groups

import (
	"context"
	"fmt"

	"github.com/coderhyme/arch-lint/internal/config"
)

type groupManager struct {
	groups map[string]Group
}

func NewGroupManager(cfg *config.Config) (GroupManager, error) {
	gm := &groupManager{
		groups: make(map[string]Group),
	}

	ctx := context.Background()

	// Phase 1: Create placeholder groups (just paths, no dependencies)
	for name, groupConfig := range cfg.Groups {
		gm.groups[name] = &groupWithRules{
			name:         name,
			pathMatchers: buildPathMatchers(groupConfig.Paths),
		}
	}

	// Phase 2: Build dependency matchers (now all groups exist)
	for name, groupConfig := range cfg.Groups {
		group, err := newGroup(ctx, name, groupConfig, gm)
		if err != nil {
			return nil, fmt.Errorf("failed to build group %s: %w", name, err)
		}
		gm.groups[name] = group
	}

	return gm, nil
}

func (gm *groupManager) GetGroups(ctx context.Context, path string) ([]Group, error) {
	var result []Group
	for _, group := range gm.groups {
		if group.MatchPath(path) {
			result = append(result, group)
		}
	}
	return result, nil
}

func (gm *groupManager) GetGroup(ctx context.Context, name string) (Group, error) {
	if group, exists := gm.groups[name]; exists {
		return group, nil
	}
	return nil, fmt.Errorf("group %s not found", name)
}
