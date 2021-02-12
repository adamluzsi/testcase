# testing double: Fake

Fakes are suppliers that have working implementations, but not the same as the production one.
Usually, they take some shortcut and have simplified version of production code.
The proper fake implementation also compliant with the [contract](/docs/contracts.md) a role interface has, like the production one.
 
An example of this shortcut can be an in-memory implementation of a Repository role interface.
This fake implementation will not engage an actual database,
but will use a simple collection to store data.

This approach allows us to do integration-testing of services without starting up a database and performing time-consuming requests.

Apart from testing, fake implementation can come in handy for prototyping and spikes.
We can quickly implement and run our system with an in-memory store, deferring decisions about database design.

Fakes can simplify local development when working with complex external systems.

example use cases:  
- payment system that always returns with successful payment and does the callback automatically on request.
- email verification process call verify callback instead of sending an email out.
- [in-memory database for testing](https://martinfowler.com/bliki/InMemoryTestDatabase.html)
 
PRO:
- can support easier local development both in local manual testing and in integration tests
- allows testing suite optimizations when using real components drastically increases the testing feedback loop time.

CON:
- testing with the proper implementation must be kept in sight else the differences between a Fake,
  and a Real implementation only reviled too late.
- a fake without an interface contract is a tech debt that needs manual maintenance and testing.
