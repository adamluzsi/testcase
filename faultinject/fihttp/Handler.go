package fihttp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/adamluzsi/testcase/faultinject"
)

type Handler struct {
	Next          http.Handler
	ServiceName   string
	FaultsMapping HandlerFaultsMapping
}

type HandlerFaultsMapping map[string][]faultinject.Tag

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var propagatedFaults []Fault
	for _, value := range r.Header.Values(HeaderName) {
		if faults, ok := h.parseHeader([]byte(value)); ok {
			tags, faults := h.mapFaultsToTags(faults)
			ctx = faultinject.Inject(ctx, tags...)
			propagatedFaults = append(propagatedFaults, faults...)
		}
	}
	if 0 < len(propagatedFaults) {
		ctx = context.WithValue(ctx, propagateCtxKey{}, propagatedFaults)
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

func (h *Handler) mapFaultsToTags(faults []Fault) (inject []faultinject.Tag, propagate []Fault) {
	for _, fault := range faults {
		if fault.ServiceName != h.ServiceName {
			propagate = append(propagate, fault)
			continue
		}
		if ntags, ok := h.FaultsMapping[fault.Name]; ok {
			inject = append(inject, ntags...)
		}
	}
	return
}

const HeaderName = `Fault-Inject`

type Fault struct {
	ServiceName string `json:"service_name,omitempty"`
	Name        string `json:"name"`
}

type propagateCtxKey struct{}
