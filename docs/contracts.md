<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [TDD, Role Interface and Contracts (role interface testing suite)](#tdd-role-interface-and-contracts-role-interface-testing-suite)
  - [Prerequisite](#prerequisite)
  - [Context](#context)
  - [Solution](#solution)
    - [When you reuse a Role Interface](#when-you-reuse-a-role-interface)
    - [Benefits](#benefits)
  - [Links](#links)
  - [TODO draft:](#todo-draft)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# TDD, Role Interface and Contracts (role interface testing suite)

## Prerequisite

- [Role Interface](interface/role.md)

## Context

Role Interface Contracts can be considered the descendant of the `Design By Contract` software correctness methodology.
Design By Contract uses preconditions and post-conditions to document and programmatically assert code interactions.
Design by Contract is a trademarked term of Bertrand Meyer and implemented in his Eiffel Language as assertions.
Design By Contract introduced Consumer and Supplier terms in this context from the testing point of view.

A consumer is a unit with a dependency that doesn't fit into the same SRP scope. For example, business logic has a reason to change if the PM requests it, but a repository from its dependencies will more likely change on a DBA's request.
Indirection.
The Consumer expresses its dependency with an interface representing a specific role the Consumer requires to achieve its purpose.
This indirection is called a role interface.
But an interface only verifies method signatures and not the implemented behaviour; thus, on its own, it doesn't invert the dependency chain.

The Consumer could not express assumptions about a supplier of a given role interface with an interface type alone.
If the Consumer relies on a given concrete supplier's behaviour to simplify its code, then that is a leaky abstraction, and the dependency chain is not inverted between the two.

Suppose we don't invert the dependency by explicitly defining this in a series of tests against a role interface type. In that case, the owner of the behavioural expectations will be each supplier implementation rather than the domain layer where the role interface is defined.

In practice, the confidence in replacing the supplier implementation degrades significantly. Thus you lose architecture flexibility and maintainability aspects of your project. For example, suppose you use PostgreSQL and need to migrate your application to a more scalable storage solution. In that case, you will be locked to using a pricey solution that mimics PostgreSQL's behaviour for your application layer's correctness that implicitly depends on that behaviour.

This design smell silently violates both the Single Responsibility Principle and Dependency Inversion Principle. However, it looks fine at first glance since the design smell aligns well with the Liskov Substitution Principle.

## Solution

The solution is a role interface and an interface testing suite, which is also called a Contract.
By creating a role interface, and an interface testing suite that tests against the role interface type,
you can clearly define your expectations from the Consumer side
and import them to the Supplier testing suite to ensure the Supplier implements the expected behaviour.

The testing subject of a contract is always a role interface.
Testing against the role interface forces the test writer to focus on the behaviour rather than any implementation details, as we don't know who will fulfil our expectations with their implementation.
Having this separation also forces a form of black-box testing through the Contract.

The Role Interface's Contract must describe all the assumptions about the Supplier's behaviour that the Consumer actively uses to simply its code.

The easiest solution is to make a struct with all the test requirements as function fields.

```go
type RoleInterfaceContract struct {
   Subject func(testing.TB) mypkg.RoleInterfaceName
   MakeXY  func(testing.TB) mypkg.XY
}
```

Then you can define an entry point for testing with a `.Test` function.
In this Test, you should define your Consumer's expectations as tests.

```go
func (c RoleInterfaceContract) Test(t *testing.T) {
   /* expectations tested here */
}
```

Optionally, you could also have a `Benchmark` function to make it easier to A/B test different Suppliers. For example, if the Consumer has performance concerns regarding an operation, the Contract should express this in the `Benchmark` function.
This approach can also help future developers to easily A/B test supplier implementations
or upgrade existing implementations where needed.


You might not have all the required interactions on the role interface to make your behavioural tests; you can fix this by defining an interface next to the Contract that embeds the testing subject role interface and requesting additional expectations from the Supplier.

```go
type RoleInterfaceContract struct {
   Subject func(testing.TB) RoleInterfaceContractSubject
   MakeXY  func(testing.TB) mypkg.XY
}

type RoleInterfaceContractSubject interface {
   mypkg.RoleInterfaceName
   FindByID(ctx context.Context, id string) (mypkg.XY, bool, error)
   DeleteByID(ctx context.Context, id string) error
}
```

Using a Contract should ensure proper boundaries for SRP scopes and non-leaky usage of dependency injection.

`testcase`'s convention to define a role interface contract is a struct that implements [`testcase#Suite`](https://pkg.go.dev/go.llib.dev/testcase#Suite)
under a `contracts` subpackage of a given domain package.
Using a different package ensures that the production code doesn't load the `testing` package into runtime because of the `*testing.T` references.
The contracts package must be under the domain package where the Consumer and its role interface are defined.

```
.
└── mypkg
    └── contracts
        └── theRoleInterfaceName.go  
```

### When you reuse a Role Interface

You can easily find yourself having role interfaces that you need to reuse in another domain package.
Or you might want to stick with a particular convention formalised by a role interface.

You can make a common interface in a separate package, and this package would own a generic expectation towards the Suppliers. Then, suppose the domain that uses this common interface requires further guarantees. In that case, they can import the common interface's Contract into their Contract and add additional test cases into their Contract.

### Benefits

If we need, we can make testing double fakes that supplies the same behaviour as the actual Suppliers but make our testing suite much less flaky while probably performing even faster.

- using fakes instead of mocks becomes possible to improve testing's feedback loop.
- dependency inversion principle not just at the static code level but at the software architecture level.
- domain logic belongs wholly to the domain context boundary.
- long-term maintenance cost

## Links

- [Role Interface by Martin Fowler](https://martinfowler.com/bliki/RoleInterface.html)
- [Design by Contract and Assertions from Eiffel Language](https://www.eiffel.org/doc/solutions/Design_by_Contract_and_Assertions)

## TODO draft:

- [ ] Making sure to define acronyms before using them (SRP jumped out to me here)
- [ ] clarify the target audience for the article and ensure no curse of knowledge here.
- [ ] reduce the meta feeling of the article by providing incremental steps in learning.
- [ ] Create a working example that starts to use this in an incremental growth style.
- [ ] introduce the distinction between header interfaces and role interfaces
  * extract this into its separate document
  * Optionally, migrate [this to test case project](https://github.com/adamluzsi/design/tree/master/interface/role-vs-header)
- [ ] replace TL;DR ambiguous parts with clean, pragmatic points
- [ ] check out the sources:
  * https://blog.thecodewhisperer.com/permalink/getting-started-with-contract-tests
  * http://jmock.org/oopsla2004.pdf
- [ ] mention fakes that we can make with contracts as an optimisation
