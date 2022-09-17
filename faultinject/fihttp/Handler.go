package fihttp

import (
	"context"
	"encoding/json"
	"net/http"
)

type Handler struct {
	Next          http.Handler
	ServiceName   string
	FaultsMapping FaultsMapping
}

type FaultsMapping map[string]InjectFn
type InjectFn func(context.Context) context.Context

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var propagatedFaults []Fault
	for _, value := range r.Header.Values(Header) {
		if faults, ok := h.parseHeader([]byte(value)); ok {
			res := h.mapFaultsToTags(faults)
			for _, injectFn := range res.Injects {
				ctx = injectFn(ctx)
			}
			propagatedFaults = append(propagatedFaults, res.Propagates...)
		}
	}
	if 0 < len(propagatedFaults) {
		ctx = Propagate(ctx, propagatedFaults...)
	}
	h.Next.ServeHTTP(w, r.WithContext(ctx))
}

func (h Handler) parseHeader(data []byte) ([]Fault, bool) {
	var faults []Fault
	if err := json.Unmarshal(data, &faults); err == nil {
		return faults, true
	}
	var fault Fault
	if err := json.Unmarshal(data, &fault); err == nil {
		return []Fault{fault}, true
	}
	return nil, false
}

type mappingResults struct {
	Injects    []InjectFn
	Propagates []Fault
}

func (h *Handler) mapFaultsToTags(faults []Fault) mappingResults {
	var mr mappingResults
	for _, fault := range faults {
		if fault.ServiceName != h.ServiceName {
			mr.Propagates = append(mr.Propagates, fault)
			continue
		}
		if inject, ok := h.FaultsMapping[fault.Name]; ok {
			mr.Injects = append(mr.Injects, inject)
		}
	}
	return mr
}
