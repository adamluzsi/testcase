<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Why to Test?](#why-to-test)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Why to Test?

The primary goal of testing is increase the speed at you gain feedback on the current system design. 
Usually when something is difficult to test, it is a good sign,
that the design might suffer from coupling, violates design principles.
In the below guide, each convention of the framework will be explained what design problem it tries to indicate.

The secondary goal of testing is to decouple the workflow into smaller units,
thus increase the efficiency by reducing the need for high mental model capacity.
As a bonus, working with tests allow you to be more resilient against random interruptions,
or get back faster to an older code base which you didn't touch since long time if ever.      

As a side effect of these conventions, the project naturally gain greater architecture flexibility and maintainability.
The lack of these aspects often credited for projects labelled as "legacy".