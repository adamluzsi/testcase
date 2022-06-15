package faultinject

type Fault struct {
	OnFunc string
	OnTag  string
	Error  error
}

func nextFault(fs *[]Fault, filter func(Fault) bool) (Fault, bool) {
	var (
		nfs   = make([]Fault, 0, len(*fs))
		fault Fault
		ok    bool
	)
	for _, f := range *fs {
		if !ok && filter(f) {
			fault = f
			ok = true
			continue
		}
		nfs = append(nfs, f)
	}
	if ok {
		*fs = nfs
	}
	return fault, ok
}
