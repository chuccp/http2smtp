package webui

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/service"
)

// fillNames stamps the owner's display name onto list rows for the admin view, going
// through the user service so the lookup (and its miss-tolerance) is shared.
func fillNames[T any](us *service.UserService, rows []T, idOf func(T) uint, set func(T, string)) {
	ids := make([]uint, 0, len(rows))
	for _, r := range rows {
		if idOf(r) > 0 {
			ids = append(ids, idOf(r))
		}
	}
	us.FillUserNames(ids, func(uid uint, name string) {
		for _, r := range rows {
			if idOf(r) == uid {
				set(r, name)
			}
		}
	})
}

// PageLink is one page number in the pager.
type PageLink struct {
	Href    string
	Label   string
	Current bool
	Gap     bool
}

// Pager renders the page-number row under a list. Filter values ride along in every
// link so paging never drops the current search.
type Pager struct {
	Total      int64
	PageNo     int
	PageSize   int
	TotalPages int
	From, To   int
	Prev       *PageLink
	Next       *PageLink
	Links      []PageLink
}

// pageLinkTo builds a link to `base` with the given filters and page number.
func pageLinkTo(base string, q url.Values, pageNo int) string {
	v := url.Values{}
	for k, vals := range q {
		for _, val := range vals {
			if val != "" && k != "pageNo" {
				v.Add(k, val)
			}
		}
	}
	if pageNo > 1 {
		v.Set("pageNo", strconv.Itoa(pageNo))
	}
	s := v.Encode()
	if s == "" {
		return base
	}
	return base + "?" + s
}

// newPager lays out the page row: first/last always visible when there is room,
// a window of ±2 around the current page, gaps marked between.
func newPager(pageNo, pageSize int, total int64, base string, q url.Values) *Pager {
	if pageSize < 1 {
		pageSize = 10
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}
	if pageNo > totalPages {
		pageNo = totalPages
	}
	if pageNo < 1 {
		pageNo = 1
	}
	p := &Pager{Total: total, PageNo: pageNo, PageSize: pageSize, TotalPages: totalPages}
	p.From = (pageNo-1)*pageSize + 1
	p.To = pageNo * pageSize
	if int64(p.To) > total {
		p.To = int(total)
	}
	if total == 0 {
		p.From = 0
	}

	appendLink := func(no int, current bool) {
		p.Links = append(p.Links, PageLink{Href: pageLinkTo(base, q, no), Label: strconv.Itoa(no), Current: current})
	}
	appendGap := func() {
		p.Links = append(p.Links, PageLink{Gap: true, Label: "…"})
	}

	lastAppended := 0
	for no := 1; no <= totalPages; no++ {
		near := no >= pageNo-2 && no <= pageNo+2
		edge := no == 1 || no == totalPages
		if !(near || edge) {
			continue
		}
		if lastAppended != 0 && no > lastAppended+1 {
			appendGap()
		}
		appendLink(no, no == pageNo)
		lastAppended = no
	}
	if pageNo > 1 {
		p.Prev = &PageLink{Href: pageLinkTo(base, q, pageNo-1), Label: "‹"}
	}
	if pageNo < totalPages {
		p.Next = &PageLink{Href: pageLinkTo(base, q, pageNo+1), Label: "›"}
	}
	return p
}

// queryOf collects the named filter fields off the request into url.Values for pager
// links. Values stay as submitted; empty entries are dropped by pageLinkTo.
func queryOf(req *web.Request, keys ...string) url.Values {
	q := url.Values{}
	for _, k := range keys {
		if v := req.Query(k); v != "" {
			q.Set(k, v)
		}
	}
	return q
}

// listPage extracts pagination for a list handler.
func listPage(req *web.Request) *web.Page {
	page, err := req.Page()
	if err != nil || page == nil {
		page = &web.Page{PageNo: 1, PageSize: 10}
	}
	if page.PageNo < 1 {
		page.PageNo = 1
	}
	if page.PageSize < 1 || page.PageSize > 200 {
		page.PageSize = 10
	}
	return page
}

// boolQuery reads a checkbox-ish query parameter ("true" when the box was checked).
func boolQuery(req *web.Request, key string) bool {
	return req.Query(key) == "true"
}

// idParam parses a numeric path parameter.
func idParam(req *web.Request, key string) (uint, error) {
	s := req.Param(key)
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", key, s)
	}
	return uint(n), nil
}
