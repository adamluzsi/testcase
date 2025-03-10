package random

import "go.llib.dev/testcase/internal"

// Name
//
// Deprecated: please use Random.Contact instead
func (r *Random) Name() contactGenerator {
	return contactGenerator{Random: r}
}

// First
//
// Deprecated: please use Contact.FirstName from Random.Contact instead
func (cg contactGenerator) First(opts ...internal.ContactOption) string {
	return cg.first(internal.ToContactConfig(opts...))
}

// Last
//
// Deprecated: please use Contact.LastName from Random.Contact instead
func (cg contactGenerator) Last() string {
	return cg.last()
}

// Email
//
// Deprecated: please use Contact.Email from Random.Contact instead
func (r *Random) Email() string {
	ng := contactGenerator{Random: r}
	return ng.email(ng.first(internal.ToContactConfig()), ng.last())
}

// SliceElement
//
// Deprecated: use random.Random#Pick instead
func (r *Random) SliceElement(slice any) any { return r.Pick(slice) }
