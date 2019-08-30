<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [The Steps struct based approach](#the-steps-struct-based-approach)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# The Steps struct based approach

Steps is an easier approach, that allows you to work with vanilla testing pkg T.Run idiom.
It builds on the foundation of variable scoping.
If you use it for setting up variables for your test cases,
you should be aware, that for that purpose, you can only execute your test cases in sequence.

```go
func TestSomething(t *testing.T) {
    var value string

    var steps = testcase.Steps{}
    t.Run(`on`, func(t *testing.T) {
        steps := steps.Before(func(t *testing.T) func() { value = "1"; return func() {} })

        t.Run(`each`, func(t *testing.T) {
            steps := steps.Before(func(t *testing.T) func() { value = "2"; return func() {} })

            t.Run(`nested`, func(t *testing.T) {
                steps := steps.Before(func(t *testing.T) func() { value = "3"; return func() {} })

                t.Run(`layer`, func(t *testing.T) {
                    steps := steps.Before(func(t *testing.T) func() { value = "4"; return func() {} })

                    t.Run(`it will setup and break down the right context`, func(t *testing.T) {
                        steps.Setup(t)

                        require.Equal(t, "4", value)
                    })
                })

                t.Run(`then`, func(t *testing.T) {
                    steps.Setup(t)

                    require.Equal(t, "3", value)
                })
            })

            t.Run(`then`, func(t *testing.T) {
                steps.Setup(t)

                require.Equal(t, "2", value)
            })
        })

        t.Run(`then`, func(t *testing.T) {
            steps.Setup(t)

            require.Equal(t, "1", value)
        })
    })
}
```
