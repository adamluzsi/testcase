<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


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
	assert.Must(t).Nil( err)
	defer storage.Close()

	err = MyStruct{Storage: s}.MyFunc()

	assert.Must(t).NotNil(err)
}

func TestMyStruct_MyFunc_storageHasActiveUser(t *testing.T) {
	u := User{}
	u.IsActive = true

	s, err := GetStorageFromENV()
	assert.Must(t).Nil( err)
	defer storage.Close()
	assert.Must(t).Nil( s.Save(&u))
	defer s.Delete(&u)

	// assert
	err = MyStruct{Storage: s}.MyFunc()
	assert.Must(t).Nil( err)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMyStruct_MyFunc_storageOnlyHasInactiveUser(t *testing.T) {
	u := User{}
	u.IsActive = false

	s, err := GetStorageFromENV()
	assert.Must(t).Nil( err)
	defer storage.Close()
	assert.Must(t).Nil( s.Save(&u))
	defer s.Delete(&u)
	

	err = MyStruct{Storage: s}.MyFunc()
	assert.Must(t).NotNil(err)
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
