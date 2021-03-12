<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Interface: indirection and/or abstraction](#interface-indirection-andor-abstraction)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Interface: indirection and/or abstraction

An interface can be thought of as a **static contract** between two components.
The owner and user of the interface called consumer, 
and the code that implements the methods of the interface is called the supplier.
At its basics, an interface represents indirection.
The interface expresses a list of possible interactions/methods with the possible implementation.

When an interface only has one real implementation supplied,
and don't abstract away implementation details
then we usually talk about a header-interface.
An example of this is when the interface talks about executing Queries.
This indirection's primary goal is to test rainy test cases, which otherwise would be difficult to recreate in a testing suite.
By providing a header interface as a dependency, we can supply a test double that can return with errors.
> [example header-interface](/docs/examples/header/main.go)

When the interface focuses on a particular goal or role and intentionally exclude any implementation details,
then we usually talk about a role-interface.
An example of this is when a storage interface talks about saving, deleting, updating or finding a domain entity.
> [example role-interface](/docs/examples/role/main.go)

A common mistake with interfaces is over-used to generate mocks, and good behaviour defined as mock expectation.
These types of mocks eventually increase the manual labour and maintenance work in the project's test coverage.
As the project ages and evolves, these mock usages often not updated correctly, 
and they continue to keep mimic a not up to date behaviour.
Refactoring also becomes more difficult, 
and often tests with mocks shifts the focus from the expected behaviour
to the implementation details of the interaction with mocks.  

The most common tech debt that is often made
when interactors or suppliers replaced with mocks in a test,
to confirm happy paths in the code flow.
Those tests most likely it will end up with asserting implementation details,
rather than testing the expected behavioural outcome.  

The suggested solution is to use as real as a possible component like the actual production implementation
or a [fake testing double](/docs/testing-double/README.md#fake) that verified with a [role interface contract](/docs/contracts.md) 
when the production variant would be a bottleneck from the point of testing feedback loop speed.
