package query

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/zitadel/zitadel/internal/database"
	"github.com/zitadel/zitadel/internal/domain"
)

var (
	groupMembersQuery = regexp.QuoteMeta("SELECT" +
		" members.creation_date" +
		", members.change_date" +
		", members.sequence" +
		", members.resource_owner" +
		", members.user_id" +
		", members.group_id" +
		", members.roles" +
		", projections.login_names3.login_name" +
		", projections.users13_humans.email" +
		", projections.users13_humans.first_name" +
		", projections.users13_humans.last_name" +
		", projections.users13_humans.display_name" +
		", projections.users13_machines.name" +
		", projections.users13_humans.avatar_key" +
		", projections.users13.type" +
		", COUNT(*) OVER () " +
		"FROM projections.group_members AS members " +
		"LEFT JOIN projections.users13_humans " +
		"ON members.user_id = projections.users13_humans.user_id " +
		"AND members.instance_id = projections.users13_humans.instance_id " +
		"LEFT JOIN projections.users13_machines " +
		"ON members.user_id = projections.users13_machines.user_id " +
		"AND members.instance_id = projections.users13_machines.instance_id " +
		"LEFT JOIN projections.users13 " +
		"ON members.user_id = projections.users13.id " +
		"AND members.instance_id = projections.users13.instance_id " +
		"LEFT JOIN projections.login_names3 " +
		"ON members.user_id = projections.login_names3.user_id " +
		"AND members.instance_id = projections.login_names3.instance_id " +
		`AS OF SYSTEM TIME '-1 ms' ` +
		"WHERE projections.login_names3.is_primary = $1")
	groupMembersColumns = []string{
		"creation_date",
		"change_date",
		"sequence",
		"resource_owner",
		"user_id",
		"group_id",
		"roles",
		"login_name",
		"email",
		"first_name",
		"last_name",
		"display_name",
		"name",
		"avatar_key",
		"type",
		"count",
	}
)

func Test_GroupMemberPrepares(t *testing.T) {
	type want struct {
		sqlExpectations sqlExpectation
		err             checkErr
	}
	tests := []struct {
		name    string
		prepare interface{}
		want    want
		object  interface{}
	}{
		{
			name:    "prepareGroupMembersQuery no result",
			prepare: prepareGroupMembersQuery,
			want: want{
				sqlExpectations: mockQueries(
					groupMembersQuery,
					nil,
					nil,
				),
			},
			object: &GroupMembers{
				GroupMembers: []*GroupMember{},
			},
		},
		{
			name:    "prepareGroupMembersQuery human found",
			prepare: prepareGroupMembersQuery,
			want: want{
				sqlExpectations: mockQueries(
					groupMembersQuery,
					groupMembersColumns,
					[][]driver.Value{
						{
							testNow,
							testNow,
							uint64(20211206),
							"ro",
							"user-id",
							"group-id",
							database.TextArray[string]{"role-1", "role-2"},
							"gigi@caos-ag.zitadel.ch",
							"gigi@caos.ch",
							"first-name",
							"last-name",
							"display name",
							nil,
							nil,
							domain.UserTypeHuman,
						},
					},
				),
			},
			object: &GroupMembers{
				SearchResponse: SearchResponse{
					Count: 1,
				},
				GroupMembers: []*GroupMember{
					{
						CreationDate:       testNow,
						ChangeDate:         testNow,
						Sequence:           20211206,
						ResourceOwner:      "ro",
						UserID:             "user-id",
						GroupID:            "group-id",
						Roles:              database.TextArray[string]{"role-1", "role-2"},
						PreferredLoginName: "gigi@caos-ag.zitadel.ch",
						Email:              "gigi@caos.ch",
						FirstName:          "first-name",
						LastName:           "last-name",
						DisplayName:        "display name",
						AvatarURL:          "",
						UserType:           domain.UserTypeHuman,
					},
				},
			},
		},
		{
			name:    "prepareGroupMembersQuery machine found",
			prepare: prepareGroupMembersQuery,
			want: want{
				sqlExpectations: mockQueries(
					groupMembersQuery,
					groupMembersColumns,
					[][]driver.Value{
						{
							testNow,
							testNow,
							uint64(20211206),
							"ro",
							"user-id",
							"group-id",
							database.TextArray[string]{"role-1", "role-2"},
							"machine@caos-ag.zitadel.ch",
							nil,
							nil,
							nil,
							nil,
							"machine-name",
							nil,
							domain.UserTypeMachine,
						},
					},
				),
			},
			object: &GroupMembers{
				SearchResponse: SearchResponse{
					Count: 1,
				},
				GroupMembers: []*GroupMember{
					{
						CreationDate:       testNow,
						ChangeDate:         testNow,
						Sequence:           20211206,
						ResourceOwner:      "ro",
						UserID:             "user-id",
						GroupID:            "group-id",
						Roles:              database.TextArray[string]{"role-1", "role-2"},
						PreferredLoginName: "machine@caos-ag.zitadel.ch",
						Email:              "",
						FirstName:          "",
						LastName:           "",
						DisplayName:        "machine-name",
						AvatarURL:          "",
						UserType:           domain.UserTypeMachine,
					},
				},
			},
		},
		{
			name:    "prepareGroupMembersQuery multiple users",
			prepare: prepareGroupMembersQuery,
			want: want{
				sqlExpectations: mockQueries(
					groupMembersQuery,
					groupMembersColumns,
					[][]driver.Value{
						{
							testNow,
							testNow,
							uint64(20211206),
							"ro",
							"user-id-1",
							"group-id",
							database.TextArray[string]{"role-1", "role-2"},
							"gigi@caos-ag.zitadel.ch",
							"gigi@caos.ch",
							"first-name",
							"last-name",
							"display name",
							nil,
							nil,
							domain.UserTypeHuman,
						},
						{
							testNow,
							testNow,
							uint64(20211206),
							"ro",
							"user-id-2",
							"group-id",
							database.TextArray[string]{"role-1", "role-2"},
							"machine@caos-ag.zitadel.ch",
							nil,
							nil,
							nil,
							nil,
							"machine-name",
							nil,
							domain.UserTypeMachine,
						},
					},
				),
			},
			object: &GroupMembers{
				SearchResponse: SearchResponse{
					Count: 2,
				},
				GroupMembers: []*GroupMember{
					{
						CreationDate:       testNow,
						ChangeDate:         testNow,
						Sequence:           20211206,
						ResourceOwner:      "ro",
						UserID:             "user-id-1",
						GroupID:            "group-id",
						Roles:              database.TextArray[string]{"role-1", "role-2"},
						PreferredLoginName: "gigi@caos-ag.zitadel.ch",
						Email:              "gigi@caos.ch",
						FirstName:          "first-name",
						LastName:           "last-name",
						DisplayName:        "display name",
						AvatarURL:          "",
						UserType:           domain.UserTypeHuman,
					},
					{
						CreationDate:       testNow,
						ChangeDate:         testNow,
						Sequence:           20211206,
						ResourceOwner:      "ro",
						UserID:             "user-id-2",
						GroupID:            "group-id",
						Roles:              database.TextArray[string]{"role-1", "role-2"},
						PreferredLoginName: "machine@caos-ag.zitadel.ch",
						Email:              "",
						FirstName:          "",
						LastName:           "",
						DisplayName:        "machine-name",
						AvatarURL:          "",
						UserType:           domain.UserTypeMachine,
					},
				},
			},
		},
		{
			name:    "prepareGroupMembersQuery sql err",
			prepare: prepareGroupMembersQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					groupMembersQuery,
					sql.ErrConnDone,
				),
				err: func(err error) (error, bool) {
					if !errors.Is(err, sql.ErrConnDone) {
						return fmt.Errorf("err should be sql.ErrConnDone got: %w", err), false
					}
					return nil, true
				},
			},
			object: (*GroupMembership)(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPrepare(t, tt.prepare, tt.object, tt.want.sqlExpectations, tt.want.err, defaultPrepareArgs...)
		})
	}
}