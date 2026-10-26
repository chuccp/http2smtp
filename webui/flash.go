package webui

import (
	"net/url"

	"github.com/chuccp/go-web-frame/web"
)

// flashOf reads the post-redirect message off the URL. Success carries an i18n key
// (?ok=saved, optionally &n= for {count}-style values); failure carries the server's
// error text verbatim (?err=...), matching what the JSON API reports.
func flashOf(req *web.Request) Flash {
	return Flash{
		OK:  req.Query("ok"),
		OKN: req.Query("n"),
		Err: req.Query("err"),
	}
}

// flashText renders the OK flash into the active language.
func flashText(f Flash, tr *Translator) string {
	if f.OK == "" {
		return ""
	}
	if f.OKN != "" {
		return tr.T(f.OK, f.OKN)
	}
	return tr.T(f.OK)
}

// backTo builds a redirect to `path` carrying the given flash as query parameters.
func backTo(path, okKey, okN, errText string) string {
	if okKey != "" {
		path += "?ok=" + okKey
		if okN != "" {
			path += "&n=" + okN
		}
	}
	if errText != "" {
		// The error text is the service's own message; it goes through query escaping
		// so a quote or an ampersand in it cannot break the URL.
		if okKey != "" {
			path += "&"
		} else {
			path += "?"
		}
		path += "err=" + url.QueryEscape(errText)
	}
	return path
}
