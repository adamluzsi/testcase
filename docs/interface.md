# Interface: indirection and/or abstraction

An interface can be thought of as a **static contract** between two components.
The owner and user of the interface called consumer, 
and the code that implements the methods of the interface is called supplier.
At its basics, an interface represent indirection.
The interface express a list of possible interactions/methods with the possible implementation.

When an interface only have one real implementation supplied,
and don't abstract away implementation details
then we usually talk about a header-interface.
An example to this is when the interface talks about executing Queries.
The main goal of this indirection is to test rainy test cases which otherwise would be difficult to recreate in a testing suite.
By providing a header interface as a dependency we can supply a test double that can return with errors.
> [example header-interface](/docs/examples/header/main.go)

When the interface focuses on a certain goal or role and intentionally exclude any implementation details,
then we usually talk about a role-interface.
An example to this is when a storage interface talks about saving, deleting, updating or finding a domain entity.
> [example role-interface](/docs/examples/role/main.go)

A common mistake with interfaces is when it is being over-used to generate mocks and good behavior defined as mock expectation.
These types of mocks eventually increase the manual labour and maintenance work in the project's test coverage.
As the project ages and evolves, these mock usages often not updated properly, 
and they continue to keep mimic a not up to date behavior.
Refactoring also becomes more difficult, 
and often tests with mocks shifts the focus from the expected behavior
to the implementation details of the interaction with mocks.  

The most common tech debt that is often made is when interactors or suppliers replaced mocks in business logic tests.
The suggested solution to that is to use as real as possible component like the actual implementation
or a [fake testing double](/docs/testing-double/fake.md) when real component is a bottleneck for the testing feedback loop.
