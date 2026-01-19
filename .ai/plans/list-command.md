# Arbor List Command - Implementation Plan

## Overview

Add `arbor list` command to display all worktrees in a repository with merge status, current worktree indicator, and main branch highlighting. The output format supports human-readable table (default), JSON (for picklist integration), and porcelain (machine-parseable).

## Command Specification

### `arbor list`

Displays all worktrees with their status.

**Flags:**
| Flag | Type | Default | Purpose |
|------|------|---------|---------|
| `--json` | bool | false | Output as JSON array (for picklist integration) |
| `--porcelain` | bool | false | Machine-parseable single-line format |
| `--sort-by` | string | `name` | Sort by: `name`, `branch`, `created` |
| `--reverse` | bool | false | Reverse sort order (for picklist toggle with `R` key) |

**Exit Codes:**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (not in a worktree, git operation failed) |

### Output Formats

**Default (Human-Readable Table):**
```
WORKTREE        BRANCH                  STATUS
main            main                    [current] [main]
feature-auth    feature/auth            [not merged]
bugfix-123      bugfix/issue-123        [merged]
```

**JSON Output (for picklist integration):**
```json
[
  {
    "path": "/Users/user/project/main",
    "branch": "main",
    "isMain": true,
    "isCurrent": true,
    "isMerged": true
  },
  {
    "path": "/Users/user/project/feature-auth",
    "branch": "feature/auth",
    "isMain": false,
    "isCurrent": false,
    "isMerged": false
  }
]
```

**Porcelain Output:**
```
/Users/user/project/main main main current merged
/Users/user/project/feature-auth feature/auth - current -
/Users/user/project/bugfix-123 bugfix/issue-123 - - merged
```

## Data Structure Enhancement

### Enhanced `Worktree` Struct

Location: `internal/git/worktree.go`

```go
type Worktree struct {
    Path      string // Full path to worktree directory
    Branch    string // Branch name (e.g., "feature/auth")
    IsMain    bool   // True if this is the main/default branch worktree
    IsCurrent bool   // True if this is the currently checked-out worktree
    IsMerged  bool   // True if branch is merged into default branch
}
```

### New Sorting Function

Location: `internal/git/worktree.go`

```go
func SortWorktrees(worktrees []Worktree, by string, reverse bool) []Worktree
```

**Sort options:**
- `name` - Sort by folder name (default)
- `branch` - Sort by branch name
- `created` - Sort by modification time

## Integration with Existing Commands

### `arbor remove`

The list output shows folder names that map directly to `arbor remove`:

```
$ arbor list
WORKTREE        BRANCH                  STATUS
feature-auth    feature/auth            [not merged]
main            main                    [current] [main] [merged]

$ arbor remove feature-auth
✓ Removes feature-auth worktree
```

### Picklist Integration

The JSON output format is designed for future TUI picklist integration:

```go
type PicklistItem struct {
    ID        string `json:"id"`        // Folder name for arbor remove
    Label     string `json:"label"`     // Display label
    Branch    string `json:"branch"`
    Path      string `json:"path"`
    IsMain    bool   `json:"isMain"`
    IsCurrent bool   `json:"isCurrent"`
    IsMerged  bool   `json:"isMerged"`
}
```

## Implementation Checklist

### Phase 1: Core Data Structures
- [x] 1.1 Modify `internal/git/worktree.go` - Add `IsMain`, `IsCurrent`, `IsMerged` fields to `Worktree` struct
- [x] 1.2 Add `SortWorktrees` function with name, branch, and created sort options
- [x] 1.3 Add `IsMerged` lookup when listing worktrees
- [x] 1.4 Add `IsCurrent` detection by comparing with current working directory
- [x] 1.5 Add `IsMain` detection by comparing with default branch

### Phase 2: Git Package Tests
- [x] 2.1 Add tests in `internal/git/worktree_test.go` for new Worktree fields
- [x] 2.2 Test `IsMerged` detection for merged and unmerged branches
- [x] 2.3 Test `IsCurrent` detection for current worktree
- [x] 2.4 Test `IsMain` detection for main/default branch
- [x] 2.5 Test `SortWorktrees` with all sort options
- [x] 2.6 Test `SortWorktrees` with reverse flag

### Phase 3: CLI Command Implementation
- [x] 3.1 Create `internal/cli/list.go` with cobra.Command structure
- [x] 3.2 Implement `--json` flag for JSON output format
- [x] 3.3 Implement `--porcelain` flag for machine-parseable output
- [x] 3.4 Implement `--sort-by` flag with name, branch, created options
- [x] 3.5 Implement `--reverse` flag for reverse sort order
- [x] 3.6 Implement default table output format
- [x] 3.7 Register command in root.go init function

### Phase 4: CLI Command Tests
- [x] 4.1 Create `internal/cli/list_test.go`
- [x] 4.2 Test empty worktree list handling
- [x] 4.3 Test shows merge status correctly
- [x] 4.4 Test marks current worktree
- [x] 4.5 Test highlights main worktree
- [x] 4.6 Test JSON output format and structure
- [x] 4.7 Test porcelain output format
- [x] 4.8 Test sorting with `--sort-by` and `--reverse`

### Phase 5: Verification
- [x] 5.1 Run full test suite: `go test ./... -v`
- [x] 5.2 Run linting: `golangci-lint run ./...`
- [x] 5.3 Manual testing of all output formats
- [x] 5.4 Verify integration with `arbor remove`

## File Changes Summary

| File | Action | Purpose |
|------|--------|---------|
| `internal/git/worktree.go` | Modify | Add IsMain, IsCurrent, IsMerged fields; Add SortWorktrees function |
| `internal/git/worktree_test.go` | Modify | Add tests for new Worktree fields and sorting |
| `internal/cli/list.go` | Create | List command implementation |
| `internal/cli/list_test.go` | Create | Command tests |
| `internal/cli/root.go` | Modify | Register list command |

## Testing Strategy

### Unit Tests (internal/git/worktree_test.go)
- `TestWorktree_IsMerged` - Merge status detection
- `TestWorktree_IsCurrent` - Current worktree detection
- `TestWorktree_IsMain` - Main branch detection
- `TestSortWorktrees_Name` - Sort by name
- `TestSortWorktrees_Branch` - Sort by branch
- `TestSortWorktrees_Created` - Sort by creation time
- `TestSortWorktrees_Reverse` - Reverse sort

### Integration Tests (internal/cli/list_test.go)
- `TestListCommand_EmptyWorktreeList` - Handles no worktrees gracefully
- `TestListCommand_ShowsMergeStatus` - Correctly reports merged/not merged
- `TestListCommand_MarksCurrentWorktree` - Identifies currently checked-out worktree
- `TestListCommand_HighlightsMainWorktree` - Marks main branch worktree
- `TestListCommand_JSONOutput` - Valid JSON structure for picklist
- `TestListCommand_PorcelainOutput` - Machine-parseable format
- `TestListCommand_Sorting` - Sorts correctly with flags
- `TestListCommand_ErrorHandling` - Handles errors (not in worktree)

## Design Decisions

### 1. JSON Output for Picklist Integration
The JSON output is designed to be easily parsed by a TUI picklist library:
- Flat structure with no nested objects
- Boolean flags for quick filtering (show only unmerged, etc.)
- Full paths included for potential future use

### 2. Sort Strategy
The `SortWorktrees` function enables:
- Alphabetical by folder name (default, matches user mental model)
- By branch name (useful for finding related features)
- By creation time (useful for finding recently active worktrees)
- Reverse toggle for picklist navigation with `R` key

### 3. Status Indicators
Status indicators in table output:
- `[current]` - The currently checked-out worktree
- `[main]` - The main/default branch worktree
- `[merged]` - Branch is merged into default branch
- `[not merged]` - Branch is not merged (default, omitted when merged)

## Future Enhancements

### TUI Picklist Integration
The JSON output format is designed for future TUI picklist:

```go
// Future picklist usage
worktrees := getWorktreesJSON()
items := []fzf.Item{
    {Title: wt.Branch, Desc: wt.Path, ...},
}
```

### Additional Sort Options
Potential future sort options:
- `modified` - Sort by last modification time
- `status` - Group by merge status (merged/unmerged)

## Dependencies

- Existing: `git.ListWorktrees(barePath)` already returns `[]Worktree`
- Existing: `git.IsMerged(barePath, branch, target)` already exists
- Existing: `git.FindBarePath(cwd)` already exists
- New: Sorting function
- New: Enhanced Worktree struct with metadata

## Rollout Plan

1. **Day 1**: Implement data structure changes and git package tests
2. **Day 2**: Implement CLI command and integration tests
3. **Day 3**: Full testing, linting, and manual verification
4. **Day 4**: Documentation updates and commit

## Success Criteria

- [ ] All tests pass (`go test ./...`)
- [ ] Linting passes (`golangci-lint run ./...`)
- [ ] Command outputs correct information for all formats
- [ ] Integration with `arbor remove` works correctly
- [ ] JSON output is valid and parseable
- [ ] Picklist-ready structure for future TUI
