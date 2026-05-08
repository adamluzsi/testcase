package testcase_test

import (
	"os"
	"strings"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/sandbox"
)

func ExampleWIP() {
	var tb testing.TB

	testcase.WIP(tb) // skip unless WIP flag is set
}

func ExampleWIP_skipForOtherUsers() {
	var tb testing.TB

	testcase.WIP(tb, "my-username", "coworker-username") // will run only for the given users
}

func TestWIP(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe("WIP", func(s *testcase.Spec) {
		var (
			doubleTB = let.Var(s, func(t *testcase.T) *doubles.TB {
				return &doubles.TB{}
			})
			users = let.VarOf[[]string](s, nil)
		)
		// act runs WIP with no users in a sandbox and stores the result in tb.
		act := func(t *testcase.T) sandbox.RunOutcome {
			return sandbox.Run(func() {
				testcase.WIP(doubleTB.Get(t), users.Get(t)...)
			})
		}

		var WhenNOWIPIsSet = func(s *testcase.Spec) {
			s.When("NOWIP is set - no wip tests / skip wip tests", func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					testcase.SetEnv(t, "NOWIP", random.Pick(t.Random, "", "T", "true", "1"))
				})

				s.Then("the test will be skipped", func(t *testcase.T) {
					act(t)

					assert.True(t, doubleTB.Get(t).Skipped())
				})
			})
		}

		s.Before(func(t *testcase.T) {
			testcase.UnsetEnv(t, "WIP")
		})

		s.When("WIP env is set", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				testcase.SetEnv(t, "WIP", random.Pick(t.Random,
					"True", "true", "1", "t", ""))
			})

			s.Then("will always run", func(t *testcase.T) {
				act(t)

				assert.False(t, doubleTB.Get(t).Skipped())
			})

			s.When("non matching user is provided", func(s *testcase.Spec) {
				users.Let(s, func(t *testcase.T) []string {
					return []string{"nonexistent-user"}
				})
				s.Before(func(t *testcase.T) {
					testcase.SetEnv(t, "TESTCASE_USER", "the-user")
				})

				s.Then("runs even when user does not match", func(t *testcase.T) {
					act(t)

					assert.False(t, doubleTB.Get(t).Skipped())
				})
			})
		})

		s.When("no users provided", func(s *testcase.Spec) {
			users.Let(s, func(t *testcase.T) []string {
				return nil
			})

			s.Then("test is skipped", func(t *testcase.T) {
				act(t)

				assert.True(t, doubleTB.Get(t).Skipped())
			})

			WhenNOWIPIsSet(s)
		})

		s.When("users provided", func(s *testcase.Spec) {
			userEnvVal := testcase.Let(s, func(t *testcase.T) string {
				username := t.Random.StringNWithCharset(5, strings.ToLower(random.CharsetAlpha()))
				testcase.SetEnv(t, "USER", username)
				return username
			})

			unrelatedUserVal := testcase.Let(s, func(t *testcase.T) string {
				makeUsernameFunc := func() string { return t.Random.StringNWithCharset(5, strings.ToLower(random.CharsetAlpha())) }
				return random.Unique(makeUsernameFunc,
					os.Getenv("TESTCASE_USER"),
					os.Getenv("USER"),
					os.Getenv("USERNAME"),
					os.Getenv("LOGNAME"),
					os.Getenv("UserName"),
					userEnvVal.Get(t))
			})

			s.And("user matches exactly", func(s *testcase.Spec) {
				users.Let(s, func(t *testcase.T) []string {
					return []string{userEnvVal.Get(t)}
				})

				s.Then("test is NOT skipped", func(t *testcase.T) {
					act(t)

					assert.False(t, doubleTB.Get(t).Skipped())
				})

				s.Context("but the user name name only match in a case insensitive way", func(s *testcase.Spec) {
					users.Let(s, func(t *testcase.T) []string {
						var name = userEnvVal.Get(t)
						if name == strings.ToLower(name) {
							name = strings.ToUpper(name)
						} else {
							name = strings.ToLower(name)
						}
						return []string{name}
					})

					s.Then("test is NOT skipped and username is matched", func(t *testcase.T) {
						act(t)

						assert.False(t, doubleTB.Get(t).Skipped())
					})
				})
			})

			s.And("user matches one of many", func(s *testcase.Spec) {
				users.Let(s, func(t *testcase.T) []string {
					return []string{
						"someone-else",
						userEnvVal.Get(t),
						"another-person",
					}
				})

				s.Then("test is NOT skipped", func(t *testcase.T) {
					act(t)

					assert.False(t, doubleTB.Get(t).Skipped())
				})
			})

			s.And("user does not match", func(s *testcase.Spec) {
				users.Let(s, func(t *testcase.T) []string {
					return []string{unrelatedUserVal.Get(t)}
				})

				s.Then("test is skipped", func(t *testcase.T) {
					act(t)

					assert.True(t, doubleTB.Get(t).Skipped())
				})

				WhenNOWIPIsSet(s)
			})
		})

		s.Describe("user env var resolution", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				testcase.UnsetEnv(t, "WIP")
			})

			username := let.Var(s, func(t *testcase.T) string {
				return t.Random.StringNWithCharset(5, strings.ToLower(random.CharsetAlpha()))
			})

			users.Let(s, func(t *testcase.T) []string {
				return []string{username.Get(t)}
			})

			s.When("USER env is set", func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					testcase.UnsetEnv(t, "USERNAME")
					testcase.UnsetEnv(t, "LOGNAME")
					testcase.SetEnv(t, "USER", username.Get(t))
				})

				s.Then("matches USER", func(t *testcase.T) {
					act(t)

					assert.False(t, doubleTB.Get(t).Skipped())
				})
			})

			s.When("TESTCASE_USER env is set", func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					testcase.SetEnv(t, "TESTCASE_USER", username.Get(t))
					testcase.UnsetEnv(t, "USER")
					testcase.UnsetEnv(t, "USERNAME")
					testcase.UnsetEnv(t, "LOGNAME")
				})

				s.Then("matches TESTCASE_USER", func(t *testcase.T) {
					act(t)

					assert.False(t, doubleTB.Get(t).Skipped())
				})

				s.Context("but even if USER env is set with something else", func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						testcase.SetEnv(t, "USER", "not-test-user-env")
					})

					s.Then("matching prioritise TESTCASE_USER", func(t *testcase.T) {
						act(t)

						assert.False(t, doubleTB.Get(t).Skipped())
					})
				})
			})

			s.When("USERNAME env is set (USER unset)", func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					testcase.UnsetEnv(t, "USER")
					testcase.UnsetEnv(t, "LOGNAME")
					testcase.SetEnv(t, "USERNAME", username.Get(t))
				})

				s.Then("matches USERNAME", func(t *testcase.T) {
					act(t)

					assert.False(t, doubleTB.Get(t).Skipped())
				})
			})

			s.When("LOGNAME env is set (USER/USERNAME unset)", func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					testcase.UnsetEnv(t, "USER")
					testcase.UnsetEnv(t, "USERNAME")
					testcase.SetEnv(t, "LOGNAME", username.Get(t))
				})

				s.Then("matches LOGNAME", func(t *testcase.T) {
					act(t)

					assert.False(t, doubleTB.Get(t).Skipped())
				})
			})

			s.When("no user env var is set", func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					testcase.UnsetEnv(t, "USER")
					testcase.UnsetEnv(t, "USERNAME")
					testcase.UnsetEnv(t, "LOGNAME")
				})

				s.Then("test is skipped", func(t *testcase.T) {
					act(t)

					assert.True(t, doubleTB.Get(t).Skipped())
				})

				WhenNOWIPIsSet(s)
			})
		})
	})
}
