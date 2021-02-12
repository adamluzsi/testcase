<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [TDD, Role Interface and Contracts](#tdd-role-interface-and-contracts)
  - [TL;DR](#tldr)
  - [Context](#context)
  - [Solution](#solution)
    - [When you reuse a Role Interface](#when-you-reuse-a-role-interface)
  - [Example #WIP](#example-wip)
  - [Links](#links)
  - [TODO draft:](#todo-draft)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# TDD, Role Interface and Contracts
## Prerequisite

- [Role Interface](interface/role.md)

## Context

Role Interface Contracts can be thought like the descendant of `Design By Contract` software correctness methodology.

Design By Contract uses preconditions and postconditions to document 
or programmatically assert the change in the state caused by a piece of a program. 
Design by Contract is a trademarked term of Bertrand Meyer and implemented in his Eiffel Language as assertions.

Design By Contract introduced Consumer and Supplier terms in this context from the testing point of view.

A consumer is a unit which has a dependency that doesn't fit into the same SRP scope,
thus it inverts the dependency by referencing an interface that expresses a need to a certain role the Consumer requires externally.
In short, this dependency is defined with a role interface.
But simply introducing an interface as a form of dependency doesn't invert the dependency chain.
If expectation and assumptions the Consumer have with the dependency is not defined in implementation detail-free format,
at the testing level, the dependency is not inverted, and thus we speak about a form of leaky abstraction.

Imagine that you think you successfully inverted the dependency by describing the role interface,
then you proceed to create an implementation 
with every important detail, you used in the Consumer tested in the supplier test coverage,
you might think you should be fine. 
What is difficult to spot is that each assumption you make about the Supplier in the Consumer, codebase side to simplify code complexity is a form of coupling between the Consumer and the Supplier implementation.
If you don't invert the dependency by explicitly defining this in a test that is implementation detail independent,
then you move part of the business logic of the Consumer under the test coverage of the Supplier.
In practice, this means that the confidence in replacing the supplier implementation degrades greatly,
thus you lose architecture flexibility and maintainability aspects of your project.
If you create a new supplier, you can't be 100% sure that you covered the same needs as the previous Supplier,
because these type of coupling usually really well hidden in the previous Supplier's test.

Because these expectations part of the Consumer business logic, 
replacing the Supplier without test coverage is no longer refactoring then
but system behaviour change on the Consumer's business logic side.

This type of design smell silently violates both the Single Responsibility Principle and Dependency Inversion Principle. However, it looks fine at first glance since the design smell aligns well with the Liskov Substitution Principle.

## Solution

The solution to this is role interface contracts or aka interface testing suite.
By creating a role interface contract which is a testing suite,
you can clearly define your expectations from the Consumer side
and use it on the supplier testing side.

The testing subject of a contract is always a role interface.
Every need to execute the tests is defined as part of the Contract dependencies.
The Role Interface Contract describes all the assumption about the behaviour of Supplier
that the Consumer actively uses to simply the code.

Two aspects must be covered.
Expected behavioural details of the Supplier
and optimisation aspects which are important from the Consumer point of view.

First, you need a `Test` function that asserts expected behavioural requirements from a supplier implementation.
These behavioural assumptions made by the Consumer to simplify and stabilise its code complexity.
Every time a Consumer assumes the behaviour of the role interface supplier,
it should be clearly defined with tests under this functionality.

You also need a `Benchmark` function that will help future developers to know what optimisation aspects matter for the Consumer.
Premature optimisation often leads to maintenance issues, and this function meant to help avoid unnecessary optimisations.
When you define a role interface contract, you likely know what performance aspects important for your Consumer to work correctly.
Not necessarily the SLA but more like what function is used often, and how or similarly.
Performance concerns should be expressed in the `Benchmark` function.
This approach will help future developers to easily A/B test supplier implementations
or to upgrade existing implementations where needed.
    
Using role interface contracts should ensure proper boundaries for SRP and non-leaky usage of DIP.

`testcase`'s convention to define a role interface contract is a struct that implements [`testcase#Contract`](https://pkg.go.dev/github.com/adamluzsi/testcase#Contract)
under a `contracts` subpackage of a given domain package.
The subpackage is required to avoid loading `testing` package into runtime because of the `*testing.T` references.
The contracts package must be under the domain package where the Consumer and its role interface dependencies are defined.

```
.
└── domains
    └── mydomain
        └── contracts
            └── RoleInterfaceName.go  
```

### When you reuse a Role Interface

You can easily find yourself having role interfaces that you need to reuse in much another domain package.
Or you might want to stick with a particular convention that is formalized by a role interface.

It would be best if you avoided at all cost moving the role interface next to a given concrete supplier implementation,
because that will make things a slippery slope.

Instead, it would be best if you defined a package which holds the commonly reused role interface(s),
and this new package must have its subpackage with all the contracts for each of the role interface.
This way you can avoid the transition of the role interfaces into a header interface of a concrete type,
and ensure that the Single Responsibility Principle and high cohesion principle is not violated.
This way, the new package will express a convention you use across your domain packages.

Having this separation also forces a form of black-box testing with the contracts,
and grant the possibility of introducing fake implementation later on.
More on that later in a different article.

### Benefits

- using fakes in testing instead of mocks become possible to improve testing's feedback loop.
- dependency inversion principle not just static code level but at software architecture level.
- domain logic belongs fully to the domain context boundary.
- long term maintenance cost 

## Example #WIP


## Links

- [Role Interface by Martin Fowler](https://martinfowler.com/bliki/RoleInterface.html)
- [Design by Contract and Assertions from Eiffel Language](https://www.eiffel.org/doc/solutions/Design_by_Contract_and_Assertions)

## TODO draft:

- [ ] Making sure to define acronyms before using them (SRP jumped out to me here)
- [ ] clarify the target audience is for the article and ensure no curse of knowledge in here.
- [ ] reduce the meta feeling of the article by providing incremental steps in learning.
- [ ] Create a working example that starts to use this in a incremental growth style.
- [ ] introduce the distinction between header interfaces and role interfaces
    * extract this into its own separate document 
    * optionally, migrate [this to test case project](https://github.com/adamluzsi/design/tree/master/interface/role-vs-header) 
- [ ] replace TL;DR ambiguous parts with pragmatic clean points
- [ ] check out the sources:
    * https://blog.thecodewhisperer.com/permalink/getting-started-with-contract-tests
    * http://jmock.org/oopsla2004.pdf
- [ ] mention fakes that can be made with contracts  
