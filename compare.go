package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/logrusorgru/aurora"
	yaml "gopkg.in/yaml.v2"

	"github.com/YuukiARIA/glanceable/models"
	"github.com/YuukiARIA/glanceable/stringSet"
)

var (
	colorOfCreated      = aurora.GreenFg
	colorOfDeleted      = aurora.RedFg
	colorOfRetained     = aurora.WhiteFg
	colorOfCreatedRole  = aurora.GreenFg | aurora.BrightFg | aurora.BoldFm
	colorOfDeletedRole  = aurora.RedFg | aurora.BrightFg | aurora.BoldFm
	colorOfRetainedRole = aurora.WhiteFg | aurora.BrightFg | aurora.BoldFm
	colorOfTeam         = aurora.WhiteFg | aurora.BrightFg | aurora.BoldFm
)

type compareResult struct {
	TeamName string              `yaml:"team_name"`
	Results  []compareRoleResult `yaml:"results"`
}

type compareRoleResult struct {
	RoleName       string           `yaml:"role_name"`
	Created        bool             `yaml:"created"`
	Deleted        bool             `yaml:"deleted"`
	UserIDsResult  compareIDsResult `yaml:"user_ids_result"`
	GroupIDsResult compareIDsResult `yaml:"group_ids_result"`
}

type compareIDsResult struct {
	CreatedIDs  []string `yaml:"created_ids"`
	DeletedIDs  []string `yaml:"deleted_ids"`
	RetainedIDs []string `yaml:"retained_ids"`
}

func (c compareResult) hasContent() bool {
	for _, roleResult := range c.Results {
		if roleResult.hasContent() {
			return true
		}
	}
	return false
}

func (c compareRoleResult) hasContent() bool {
	return c.UserIDsResult.hasContent() || c.GroupIDsResult.hasContent()
}

func (c compareIDsResult) hasContent() bool {
	return len(c.CreatedIDs) > 0 || len(c.DeletedIDs) > 0 || len(c.RetainedIDs) > 0
}

func (c compareResult) show() {
	fmt.Println("team: " + aurora.Colorize(c.TeamName, colorOfTeam).String())
	fmt.Println()

	for _, roleResult := range c.Results {
		if roleResult.hasContent() {
			c := getRoleColor(roleResult.Created, roleResult.Deleted)
			fmt.Println("role: " + aurora.Colorize(roleResult.RoleName, c).String())
			roleResult.show()
		}
	}
}

func (c compareRoleResult) show() {
	printIndent(2)
	fmt.Println("users:")
	if c.UserIDsResult.hasContent() {
		c.UserIDsResult.show()
	} else {
		printIndent(4)
		fmt.Println("(none)")
	}
	fmt.Println()
	printIndent(2)
	fmt.Println("groups:")
	if c.GroupIDsResult.hasContent() {
		c.GroupIDsResult.show()
	} else {
		printIndent(4)
		fmt.Println("(none)")
	}
	fmt.Println()
}

func (c compareIDsResult) show() {
	if len(c.RetainedIDs) > 0 {
		showAsRetained(4, c.RetainedIDs...)
	}
	if len(c.CreatedIDs) > 0 {
		showAsCreated(4, c.CreatedIDs...)
	}
	if len(c.DeletedIDs) > 0 {
		showAsDeleted(4, c.DeletedIDs...)
	}
}

func getRoleColor(created, deleted bool) aurora.Color {
	if created {
		return colorOfCreatedRole
	} else if deleted {
		return colorOfDeletedRole
	}
	return colorOfRetainedRole
}

func Compare(oldTeam, newTeam models.Team) compareResult {
	roleResults := make([]compareRoleResult, 0)

	for roleName, oldRule := range oldTeam.Auth {
		var roleResult compareRoleResult

		newRule, exists := newTeam.Auth[roleName]
		if exists {
			roleResult = compareRule(oldRule, newRule)
		} else {
			roleResult = compareRoleResult{
				RoleName:       roleName,
				Deleted:        true,
				UserIDsResult:  newCompareIDsResult(nil, oldRule.Users, nil),
				GroupIDsResult: newCompareIDsResult(nil, oldRule.Groups, nil),
			}
		}
		roleResults = append(roleResults, roleResult)
	}

	for roleName, newRule := range newTeam.Auth {
		_, exists := oldTeam.Auth[roleName]
		if !exists {
			roleResult := compareRoleResult{
				RoleName:       roleName,
				Created:        true,
				UserIDsResult:  newCompareIDsResult(newRule.Users, nil, nil),
				GroupIDsResult: newCompareIDsResult(newRule.Groups, nil, nil),
			}
			roleResults = append(roleResults, roleResult)
		}
	}

	return compareResult{TeamName: oldTeam.Name, Results: roleResults}
}

func ShowDefaultFormat(result compareResult) {
	result.show()
}

func ShowJSONFormat(result compareResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(&result)
}

func ShowYAMLFormat(result compareResult) error {
	bytes, err := yaml.Marshal(&result)
	if err != nil {
		return err
	}
	fmt.Print(string(bytes))
	return nil
}

func compareRule(oldRule, newRule *models.AuthRule) compareRoleResult {
	return compareRoleResult{
		RoleName:       oldRule.RoleName,
		UserIDsResult:  compareIds(oldRule.Users, newRule.Users),
		GroupIDsResult: compareIds(oldRule.Groups, newRule.Groups),
	}
}

func compareIds(oldIds, newIds []string) compareIDsResult {
	oldIdsSet, newIdsSet := stringSet.New(oldIds...), stringSet.New(newIds...)
	deletedIdsSet, createdIdsSet, retainedIdsSet := stringSet.Partition(oldIdsSet, newIdsSet)
	return newCompareIDsResult(createdIdsSet.Array(), deletedIdsSet.Array(), retainedIdsSet.Array())
}

func newCompareIDsResult(createdIDs, deletedIDs, retainedIDs []string) compareIDsResult {
	empty := make([]string, 0)
	if createdIDs == nil {
		createdIDs = empty
	}
	if deletedIDs == nil {
		deletedIDs = empty
	}
	if retainedIDs == nil {
		retainedIDs = empty
	}
	return compareIDsResult{CreatedIDs: createdIDs, DeletedIDs: deletedIDs, RetainedIDs: retainedIDs}
}

func showAsCreated(indentLevel int, values ...string) {
	showWithColor(indentLevel, "+", colorOfCreated, values...)
}

func showAsDeleted(indentLevel int, values ...string) {
	showWithColor(indentLevel, "-", colorOfDeleted, values...)
}

func showAsRetained(indentLevel int, values ...string) {
	showWithColor(indentLevel, " ", colorOfRetained, values...)
}

func showWithColor(indentLevel int, prefix string, color aurora.Color, values ...string) {
	for _, value := range values {
		printIndent(indentLevel)
		fmt.Println(aurora.Colorize(fmt.Sprintf("%s %s", prefix, value), color))
	}
}

func printIndent(n int) {
	for i := 0; i < n; i++ {
		fmt.Print(" ")
	}
}
