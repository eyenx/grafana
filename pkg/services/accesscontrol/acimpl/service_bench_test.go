package acimpl

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/accesscontrol/database"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/require"
)

const concurrency = 20
const batchSize = 1000

type bounds struct {
	start, end int
}

func concurrentBatch(workers, count, size int, eachFn func(start, end int) error) error {
	var wg sync.WaitGroup
	alldone := make(chan bool) // Indicates that all workers have finished working
	chunk := make(chan bounds) // Gives the workers the bounds they should work with
	ret := make(chan error)    // Allow workers to notify in case of errors
	defer close(ret)

	// Launch all workers
	for x := 0; x < workers; x++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ck := range chunk {
				if err := eachFn(ck.start, ck.end); err != nil {
					ret <- err
					return
				}
			}
		}()
	}

	go func() {
		// Tell the workers the chunks they have to work on
		for i := 0; i < count; {
			end := i + size
			if end > count {
				end = count
			}

			chunk <- bounds{start: i, end: end}

			i = end
		}
		close(chunk)

		// Wait for the workers
		wg.Wait()
		close(alldone)
	}()

	// wait for an error or for all workers to be done
	select {
	case err := <-ret:
		return err
	case <-alldone:
		break
	}
	return nil
}

func setupBenchEnv(b *testing.B, usersCount, resourceCount int) (accesscontrol.Service, *user.SignedInUser) {
	now := time.Now()
	sqlStore := db.InitTestDB(b)
	store := database.ProvideService(sqlStore)
	acService := &Service{
		cfg:           setting.NewCfg(),
		log:           log.New("accesscontrol-test"),
		registrations: accesscontrol.RegistrationList{},
		store:         store,
		roles:         accesscontrol.BuildBasicRoleDefinitions(),
	}

	// Prepare default permissions
	action1 := "resources:action1"
	err := acService.DeclareFixedRoles(accesscontrol.RoleRegistration{
		Role:   accesscontrol.RoleDTO{Name: "fixed:test:role", Permissions: []accesscontrol.Permission{{Action: action1}}},
		Grants: []string{string(org.RoleViewer)},
	})
	require.NoError(b, err)
	err = acService.RegisterFixedRoles(context.Background())
	require.NoError(b, err)

	// Prepare managed permissions
	action2 := "resources:action2"
	users := make([]user.User, 0, usersCount)
	orgUsers := make([]org.OrgUser, 0, usersCount)
	roles := make([]accesscontrol.Role, 0, usersCount)
	userRoles := make([]accesscontrol.UserRole, 0, usersCount)
	permissions := make([]accesscontrol.Permission, 0, resourceCount*usersCount)
	for u := 1; u < usersCount+1; u++ {
		users = append(users, user.User{
			ID:      int64(u),
			Name:    fmt.Sprintf("user%v", u),
			Login:   fmt.Sprintf("user%v", u),
			Email:   fmt.Sprintf("user%v@example.org", u),
			Created: now,
			Updated: now,
		})
		orgUsers = append(orgUsers, org.OrgUser{
			ID:      int64(u),
			UserID:  int64(u),
			OrgID:   1,
			Role:    org.RoleViewer,
			Created: now,
			Updated: now,
		})
		roles = append(roles, accesscontrol.Role{
			ID:      int64(u),
			UID:     fmt.Sprintf("managed_users_%v_permissions", u),
			Name:    fmt.Sprintf("managed:users:%v:permissions", u),
			Version: 1,
			Created: now,
			Updated: now,
		})
		userRoles = append(userRoles, accesscontrol.UserRole{
			ID:      int64(u),
			OrgID:   1,
			RoleID:  int64(u),
			UserID:  int64(u),
			Created: now,
		})

		for r := 1; r < resourceCount+1; r++ {
			permissions = append(permissions, accesscontrol.Permission{
				RoleID:  int64(u),
				Action:  action2,
				Scope:   fmt.Sprintf("resources:id:%v", r),
				Created: now,
				Updated: now,
			})
		}
	}

	// Populate store
	if errInsert := concurrentBatch(concurrency, len(roles), batchSize, func(start, end int) error {
		err := sqlStore.WithDbSession(context.Background(), func(sess *db.Session) error {
			if _, err := sess.Insert(users[start:end]); err != nil {
				return err
			}
			if _, err := sess.Insert(orgUsers[start:end]); err != nil {
				return err
			}
			if _, err := sess.Insert(roles[start:end]); err != nil {
				return err
			}
			_, err := sess.Insert(userRoles[start:end])
			return err
		})
		return err
	}); errInsert != nil {
		require.NoError(b, err, "could not insert users and roles")
		return nil, nil
	}

	if errInsert := concurrentBatch(concurrency, len(permissions), batchSize, func(start, end int) error {
		return sqlStore.WithDbSession(context.Background(), func(sess *db.Session) error {
			_, err := sess.Insert(permissions[start:end])
			return err
		})
	}); errInsert != nil {
		require.NoError(b, err, "could not insert permissions")
		return nil, nil
	}

	// Allow signed in user to view all users permissions in the worst way
	userPermissions := map[string][]string{}
	for u := 1; u < usersCount+1; u++ {
		userPermissions[accesscontrol.ActionUsersPermissionsRead] =
			append(userPermissions[accesscontrol.ActionUsersPermissionsRead], fmt.Sprintf("users:id:%v", u))
	}
	return acService, &user.SignedInUser{OrgID: 1, Permissions: map[int64]map[string][]string{1: userPermissions}}
}

func benchSearchUsersPermissions(b *testing.B, usersCount, resourceCount int) {
	acService, siu := setupBenchEnv(b, usersCount, resourceCount)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		usersPermissions, err := acService.SearchUsersPermissions(context.Background(), siu, 1, accesscontrol.SearchOptions{ActionPrefix: "resources:"})
		require.NoError(b, err)
		require.Len(b, usersPermissions, usersCount)
		for _, permissions := range usersPermissions {
			// action1 on all resource + action2
			require.Len(b, permissions, resourceCount+1)
		}
	}
}

// Lots of resources
func BenchmarkSearchUsersPermissions_10_1K(b *testing.B)  { benchSearchUsersPermissions(b, 10, 1000) }  // ~0.047s/op
func BenchmarkSearchUsersPermissions_10_10K(b *testing.B) { benchSearchUsersPermissions(b, 10, 10000) } // ~0.5s/op
func BenchmarkSearchUsersPermissions_10_100K(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}
	benchSearchUsersPermissions(b, 10, 100000)
} // ~4.6s/op
func BenchmarkSearchUsersPermissions_10_1M(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}
	benchSearchUsersPermissions(b, 10, 1000000)
} // ~55.36s/op

// Lots of users (most probable case)
func BenchmarkSearchUsersPermissions_1K_10(b *testing.B)  { benchSearchUsersPermissions(b, 1000, 10) }  // ~0.056s/op
func BenchmarkSearchUsersPermissions_10K_10(b *testing.B) { benchSearchUsersPermissions(b, 10000, 10) } // ~0.58s/op
func BenchmarkSearchUsersPermissions_100K_10(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}
	benchSearchUsersPermissions(b, 100000, 10)
} // ~6.21s/op
func BenchmarkSearchUsersPermissions_1M_10(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}
	benchSearchUsersPermissions(b, 1000000, 10)
} // ~57s/op

// Lots of both
func BenchmarkSearchUsersPermissions_10K_100(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}
	benchSearchUsersPermissions(b, 10000, 100)
} // ~1.45s/op
func BenchmarkSearchUsersPermissions_10K_1K(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}
	benchSearchUsersPermissions(b, 10000, 1000)
} // ~50s/op

// Benchmarking search when we specify Action and Scope
func benchSearchUsersWithPerm(b *testing.B, usersCount, resourceCount int) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}
	acService, siu := setupBenchEnv(b, usersCount, resourceCount)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		usersPermissions, err := acService.SearchUsersPermissions(context.Background(), siu, 1,
			accesscontrol.SearchOptions{Action: "resources:action2", Scope: "resources:id:1"})
		require.NoError(b, err)
		require.Len(b, usersPermissions, usersCount)
		for _, permissions := range usersPermissions {
			require.Len(b, permissions, 1)
		}
	}
}

func BenchmarkSearchUsersWithPerm_1K_10(b *testing.B)   { benchSearchUsersWithPerm(b, 1000, 10) }     // ~0.045s/op
func BenchmarkSearchUsersWithPerm_1K_100(b *testing.B)  { benchSearchUsersWithPerm(b, 1000, 100) }    // ~0.038s/op
func BenchmarkSearchUsersWithPerm_1K_1K(b *testing.B)   { benchSearchUsersWithPerm(b, 1000, 1000) }   // ~0.033s/op
func BenchmarkSearchUsersWithPerm_1K_10K(b *testing.B)  { benchSearchUsersWithPerm(b, 1000, 10000) }  // ~0.033s/op
func BenchmarkSearchUsersWithPerm_1K_100K(b *testing.B) { benchSearchUsersWithPerm(b, 1000, 100000) } // ~0.056s/op

func BenchmarkSearchUsersWithPerm_10K_10(b *testing.B)  { benchSearchUsersWithPerm(b, 10000, 10) }    // ~0.11s/op
func BenchmarkSearchUsersWithPerm_10K_100(b *testing.B) { benchSearchUsersWithPerm(b, 10000, 100) }   // ~0.12s/op
func BenchmarkSearchUsersWithPerm_10K_1K(b *testing.B)  { benchSearchUsersWithPerm(b, 10000, 1000) }  // ~0.12s/op
func BenchmarkSearchUsersWithPerm_10K_10K(b *testing.B) { benchSearchUsersWithPerm(b, 10000, 10000) } // ~1.77s/op

func BenchmarkSearchUsersWithPerm_20K_10(b *testing.B)  { benchSearchUsersWithPerm(b, 20000, 10) }    // ~0.22s/op
func BenchmarkSearchUsersWithPerm_20K_100(b *testing.B) { benchSearchUsersWithPerm(b, 20000, 100) }   // ~0.22s/op
func BenchmarkSearchUsersWithPerm_20K_1K(b *testing.B)  { benchSearchUsersWithPerm(b, 20000, 1000) }  // ~0.25s/op
func BenchmarkSearchUsersWithPerm_20K_10K(b *testing.B) { benchSearchUsersWithPerm(b, 20000, 10000) } // ~s/op

func BenchmarkSearchUsersWithPerm_100K_10(b *testing.B)  { benchSearchUsersWithPerm(b, 100000, 10) }  // ~0.88s/op
func BenchmarkSearchUsersWithPerm_100K_100(b *testing.B) { benchSearchUsersWithPerm(b, 100000, 100) } // ~0.72s/op
