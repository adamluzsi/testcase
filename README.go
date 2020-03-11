/*

Package testcase implements two approaches to help you to do nested BDD style testing in golang.



Spec Variables

in your spec, you can use the `*testcase.variables` object,
for fetching values for your objects.
Using them is gives you the ability to create value for them,
only when you are in the right testing scope that responsible
for providing an example for the expected value.

In test case scopes you will receive a structure ptr called `*testcase.variables`
which will represent values that you configured for your test case with `Let`.

Values in `*testcase.variables` are safe to use during T#Parallel.



Spec Hooks

Hooks help you setup common things for each test case.
For example clean ahead, clean up, mock expectation configuration,
and similar things can be done in hooks,
so your test case blocks with `Then` only represent the expected result(s).

In case you work with something that depends on side-effects,
such as database tests, you can use the hooks,
to create clean-ahead / clean-up blocks.

Also if you use gomock, you can use the spec#Around function,
to set up the mock with a controller, and in the teardown function,
call the gomock.Controller#Finish function,
so your test cases will be only about
what is the different behavior from the rest of the test cases.

It will panic if you use hooks or variable preparation in an ambiguous way,
or when you try to access variable that doesn't exist in the context where you do so.
It tries to panic with friendly and supportive messages, but that is highly subjective.

*/
package testcase
