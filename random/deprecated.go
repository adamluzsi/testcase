package random

import "github.com/adamluzsi/testcase/internal"

// Name
//
// DEPRECATED: please use Random.Contact instead
func (r *Random) Name() contactGenerator {
	return contactGenerator{Random: r}
}

// First
//
// DEPRECATED: please use Contact.FirstName from Random.Contact instead
func (cg contactGenerator) First(opts ...internal.ContactOption) string {
	return cg.first(internal.ToContactConfig(opts...))
}

// Last
//
// DEPRECATED: please use Contact.LastName from Random.Contact instead
func (cg contactGenerator) Last() string {
	return cg.last()
}

// Email
//
// DEPRECATED: please use Contact.Email from Random.Contact instead
func (r *Random) Email() string {
	ng := contactGenerator{Random: r}
	return ng.email(ng.first(internal.ToContactConfig()), ng.last())
}
