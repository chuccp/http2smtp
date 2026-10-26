package rest

import "github.com/chuccp/go-web-frame/web"

// maxPageSize bounds what one list request can pull into memory. The framework's
// DefaultPage only lifts values below 1, so without this a caller can ask any list
// endpoint for the whole table (and the token list supplements each row with further
// queries on top of that).
const maxPageSize = 200

// capPage clamps the page parameters extracted from a request.
func capPage(page *web.Page) *web.Page {
	if page == nil {
		return &web.Page{PageNo: 1, PageSize: 10}
	}
	if page.PageNo < 1 {
		page.PageNo = 1
	}
	if page.PageSize < 1 {
		page.PageSize = 10
	}
	if page.PageSize > maxPageSize {
		page.PageSize = maxPageSize
	}
	return page
}
