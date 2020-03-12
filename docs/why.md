<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [To what problem, this project give a solution?](#to-what-problem-this-project-give-a-solution)
  - [What does it provide on top of core `testing` package ?](#what-does-it-provide-on-top-of-core-testing-package-)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# To what problem, this project give a solution?

## What does it provide on top of core `testing` package ?

`testcase` package main goal is to provide test specification style,
where you can incrementally explain the execution context of a test edge case.
Through this you have to make a conscious decision to `test` or `skip` a certain edge case.

To provide some explanation, imagine the following test specification:

```gherkin
Given I have a User the database
When the user is active
Then The user state returned as active
```

Traditionally, for an integration test, first you have to create a user,
and you have to setup the entity to reflect the test edge case context,
and then you can persist it.
After that you can make your assertions.

In situations where you have a lot of small contextual specification details,
for example if a resource like a database has a certain state,
which affects the behavior of the component we currently test,
we have to create test specification like the following example. 

```go
func TestMyStruct_MyFunc_noUserHadBeenSavedBefore(t *testing.T) {
	s, err := GetStorageFromENV()
	require.Nil(t, err)
	defer storage.Close()

	err = MyStruct{Storage: s}.MyFunc()

	require.Error(t, err)
}

func TestMyStruct_MyFunc_storageHasActiveUser(t *testing.T) {
	u := User{}
	u.IsActive = true

	s, err := GetStorageFromENV()
	require.Nil(t, err)
	defer storage.Close()
	require.Nil(t, s.Save(&u))
	defer s.Delete(&u)

	// assert
	err = MyStruct{Storage: s}.MyFunc()
	require.Nil(t, err)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMyStruct_MyFunc_storageOnlyHasInactiveUser(t *testing.T) {
	u := User{}
	u.IsActive = false

	s, err := GetStorageFromENV()
	require.Nil(t, err)
	defer storage.Close()
	require.Nil(t, s.Save(&u))
	defer s.Delete(&u)
	

	err = MyStruct{Storage: s}.MyFunc()
	require.Error(t, err)
}
```

And if we want to tweak the arrange part,
we have to duplicate the rest of the parts multiple time then.

In `testcase` instead of this approach,
you start to specify with the most common part of the context,
that in terms of execution is constant for a certain context,
and then you start add more and more context in each test sub context.
And when the specification feels to have to many nesting layers,
you try to split the component into smaller logical chunks.

```go
func TestMyStruct_MyFunc(t *testing.T) {
	s := testcase.NewSpec(t)

	storage, err := GetStorageFromENV()
	require.Nil(t, err)
	defer storage.Close()

	subject := func(t *testcase.T) error {
		ms := MyStruct{Storage: t.I(`storage`).(*Storage)}
		return ms.MyFunc()
	}

	s.Let(`user`, func(t *testcase.T) interface{} {
		return &User{IsActive: t.I(`is user active?`).(bool)}
	})

	s.When(`a user is saved to the storage`, func(s *testcase.Spec) {
		s.Around(func(t *testcase.T) func() {
			u := t.I(`user`).(*User)
			require.Nil(t, storage.Save(u))
			return func() { require.Nil(t, storage.Delete(u)) }
		})

		s.And(`the at least one user is active`, func(s *testcase.Spec) {
			s.LetValue(`is user active?`, true)

			s.Then(`no error expected`, func(t *testcase.T) {
				require.Nil(t, subject(t))
			})
		})

		s.And(`all the users are inactive`, func(s *testcase.Spec) {
			s.LetValue(`is user active?`, false)

			s.Then(`error expected`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})
	})

	s.When(`no user had been saved before in the storage`, func(s *testcase.Spec) {
		s.LetValue(`is user active?`, rand.Intn(1) == 0) // to ensure input

		s.Then(`error expected`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})
}
```   
