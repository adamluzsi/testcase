/*

Package testcase implements two approaches to help you to do nested BDD style testing in golang.



Introduction

testcase framework Is meant to help the test writer focus on defining the behavior of a given testing subject,
while helping remaining disciplined with the way the test is being made.
This disciplined ways of working will later can be used to make changes and refactoring on the testing suite a breeze.
The main goal of the project is to provide long term maintainability and re-usability in testing.



Spec Variables

In your spec, you can define and then later use variables that bound to test execution.
These values lazy evaluated for each test case, and then cached till the life time of the test.
These values are thread safe, concurrently running test don't have visibility to the other test's variables.
As a benefit of this, you can use parallel execution whenever your test subject don't have side effects.

On top of that, if you need to build a certain context that will ensure the variable content,
you can reference to a variable earlier than you actually define it.
This helps you to ensure that each edge case is covered properly,
and you don't have leaking values from one test to an another.
You can learn more about this under Spec#Let Spec#LetValue and Var.



Spec Hooks

When you composite a testing context,
each testing context meant to have a description
and a corresponding code that fulfils the described context.
To do so, Spec#Let might not be enough.
For this purpose you can use hooks that will be executed Before, After or Around the Spec#Test blocks.
Each testing context inherit all the exiting hook from its parent context.
Therefore setting up tests with common events should be easy as creating a Spec#Context with a hook.
If additional events required you can do so in sub contexts.
Hooks can be also used to do common cleanups before or after each test scenario.
Defining a hook after a test assertion is not possible,
as it would ruin consistency of the testing scope.
You will receive a warning about this if you attempt to do so.

*/
package testcase
