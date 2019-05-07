/*

package testcase implements two approaches to help you to do nested BDD style testing in golang.



Spec Variables

in your spec, you can use the `*testcase.V` object,
for fetching values for your objects.
Using them is gives you the ability to create value for them,
only when you are in the right testing scope that responsible
for providing an example for the expected value.

In test case scopes you will receive a structure ptr called `*testcase.V`
which will represent values that you configured for your test case with `Let`.

Values in `*testcase.V` are safe to use during T#Parallel.



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



Yes, but why

I made a list of requirements for myself, and then looked trough the available testing frameworks in golang:
	* works perfectly well with `go test` command out of the box
	  * includes `-run` option usability
	* allow me to run one test edge case easily from the specification
	* don't build singleton objects outside of my test function scope
	* allow me to run test cases in concurrent execution for specification where I know that no side effect expected.
	  * this is especially important me, because I love quick test feedback loops
	* allow me to define variables in a way, that they receive concrete value later
	  * this help me build spec coverage, where if I forgot a edge case regarding a variable, the spec will simply panic about early on.
	* allow me to define variables that can be safely overwritten with nested scopes
	* I want to use [stretchr/testify](https://github.com/stretchr/testify), so assertions not necessary for me
	  * or more precisely, I needed something that guaranteed to allow me the usage of that pkg

While I liked the existing solutions, I felt that the way I would use them would leave out one or more point from my requirements.
So I ended up making a small design about how it would be great for me to test.
I took inspiration from [rspec](https://github.com/rspec/rspec),
as I loved the time I spent working with that framework.
This is how this pkg is made.



So what is the main difference from the others

Using this pkg allow you to set up input variables for your test subject,
in a way that the variables are belong to a certain test context scope only,
and cannot leak out to other test executions implicitly.

This will allow you to create test cases,
where if you forgot to set the context correctly,
the tests will panic early on and warn you about.

Also if you run it in parallel, there is a guarantee that your variables will not be leaked out,
and will not affect your other test cases, trough a shared variable,
because each test case execution has its own dedicated set of variables.

To me fast feedback cycle from the test is really important,
and go `*testing.T#Parallel` functionality is really liked.
And I needed a solution that would allow me to create specifications,
that are thread safe for concurrent execution.

*/
package testcase
