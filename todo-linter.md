Understandable
Tests should shout out what it is that they are testing and asserting.

Maintainable
Good tests should be easy to change, and should make it easy to change the code that they are testing too.

Repeatable
We should get the same result for the same version of the code every time we run them. So no timezone or concurrency problems.

Atomic
A good test should stand alone and not depend on arrangements it doesn't control.

Necessary
Good tests are there for a reason, they aren't randomly testing things, they express a new, different perspective on our code.

Granular
A test should assert a single outcome. We don't what to have to wade through logs to understand why a test fails, it should be crystal clear.

Fast
Good tests are fast, they run quickly enough that we are happy to run them after even tiny changes to our code. 

Simple
As well as asserting a single outcome, good tests in good systems are very simple. My favourite description is from 
@JonJagger
 "A good test has a cyclomatic complexity of 1"
