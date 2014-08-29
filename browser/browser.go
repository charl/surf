package browser

import (
	"encoding/base64"
	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/errors"
	"github.com/headzoo/surf/event"
	"github.com/headzoo/surf/jar"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Attribute represents a Browser capability.
type Attribute int

// AttributeMap represents a map of Attribute values.
type AttributeMap map[Attribute]bool

const (
	// SendRefererAttribute instructs a Browser to send the Referer header.
	SendReferer Attribute = iota

	// MetaRefreshHandlingAttribute instructs a Browser to handle the refresh meta tag.
	MetaRefreshHandling

	// FollowRedirectsAttribute instructs a Browser to follow Location headers.
	FollowRedirects
)

// LogLevel represents a logging level.
type LogLevel int8

const (
	// LogLevelDebug is the most verbose logging level.
	LogLevelDebug LogLevel = iota

	// LogLevelInfo only logs basic requests and form submissions.
	LogLevelInfo

	// LogLevelError logs errors.
	LogLevelError
)

// InitialAssetsArraySize is the initial size when allocating a slice of page
// assets. Increasing this size may lead to a very small performance increase
// when downloading assets from a page with a lot of assets.
var InitialAssetsSliceSize = 20

// Authorization holds the username and password for authorization.
type Authorization struct {
	Username string
	Password string
}

// Browsable represents an HTTP web browser.
type Browsable interface {
	event.Eventable

	// SetUserAgent sets the user agent.
	SetUserAgent(ua string)

	// SetAttribute sets a browser instruction attribute.
	SetAttribute(a Attribute, v bool)

	// SetAttributes is used to set all the browser attributes.
	SetAttributes(a AttributeMap)

	// SetAuthorization sets the username and password to use during authorization.
	SetAuthorization(username, password string)

	// SetBookmarksJar sets the bookmarks jar the browser uses.
	SetBookmarksJar(bj jar.Bookmarks)

	// SetCookieJar is used to set the cookie jar the browser uses.
	SetCookieJar(cj http.CookieJar)

	// SetHistoryJar is used to set the history jar the browser uses.
	SetHistoryJar(hj jar.History)

	// SetRecorderJar sets a jar.RecorderJar that will record browser states.
	SetRecorderJar(rj jar.Recorder)

	// SetHeadersJar sets the headers the browser sends with each request.
	SetHeadersJar(h http.Header)

	// SetEventDispatcher sets the event dispatcher.
	SetEventDispatcher(ed event.Eventable)

	// SetLogger sets the instance that will be used to log events.
	SetLogger(l *log.Logger, lev LogLevel)

	// AddRequestHeader adds a header the browser sends with each request.
	AddRequestHeader(name, value string)

	// Open requests the given URL using the GET method.
	Open(url string) error

	// OpenForm appends the data values to the given URL and sends a GET request.
	OpenForm(url string, data url.Values) error

	// OpenBookmark calls Get() with the URL for the bookmark with the given name.
	OpenBookmark(name string) error

	// Post requests the given URL using the POST method.
	Post(url string, contentType string, body io.Reader) error

	// PostForm requests the given URL using the POST method with the given data.
	PostForm(url string, data url.Values) error

	// Back loads the previously requested page.
	Back() bool

	// Reload duplicates the last successful request.
	Reload() error

	// Recorder returns the set recorder.
	Recorder() jar.Recorder

	// Bookmark saves the page URL in the bookmarks with the given name.
	Bookmark(name string) error

	// Click clicks on the page element matched by the given expression.
	Click(expr string) error

	// Form returns the form in the current page that matches the given expr.
	Form(expr string) (Submittable, error)

	// Forms returns an array of every form in the page.
	Forms() []Submittable

	// Links returns an array of every link found in the page.
	Links() []*Link

	// Images returns an array of every image found in the page.
	Images() []*Image

	// Stylesheets returns an array of every stylesheet linked to the document.
	Stylesheets() []*Stylesheet

	// Scripts returns an array of every script linked to the document.
	Scripts() []*Script

	// SiteCookies returns the cookies for the current site.
	SiteCookies() []*http.Cookie

	// ResolveUrl returns an absolute URL for a possibly relative URL.
	ResolveUrl(u *url.URL) *url.URL

	// ResolveStringUrl works just like ResolveUrl, but the argument and return value are strings.
	ResolveStringUrl(u string) (string, error)

	// Download writes the contents of the document to the given writer.
	Download(o io.Writer) (int64, error)

	// Url returns the page URL as a string.
	Url() *url.URL

	// StatusCode returns the response status code.
	StatusCode() int

	// Title returns the page title.
	Title() string

	// ResponseHeaders returns the page headers.
	ResponseHeaders() http.Header

	// Body returns the page body as a string of html.
	Body() string

	// Dom returns the inner *goquery.Selection.
	Dom() *goquery.Selection

	// Find returns the dom selections matching the given expression.
	Find(expr string) *goquery.Selection
}

// Default is the default Browser implementation.
type Browser struct {
	*event.Dispatcher

	// state is the current browser state.
	state *jar.State

	// userAgent is the User-Agent header value sent with requests.
	userAgent string

	// cookies stores cookies for every site visited by the browser.
	cookies http.CookieJar

	// bookmarks stores the saved bookmarks.
	bookmarks jar.Bookmarks

	// history stores the visited pages.
	history jar.History

	// recorder is used to record browser states and play them back.
	recorder jar.Recorder

	// headers are additional headers to send with each request.
	headers http.Header

	// attributes is the set browser attributes.
	attributes AttributeMap

	// refresh is a timer used to meta refresh pages.
	refresh *time.Timer

	// auth stores the authorization credentials.
	auth Authorization

	// logger will log events.
	logger *log.Logger

	// logLevel is the logging level.
	logLevel LogLevel
}

// Open requests the given URL using the GET method.
func (bow *Browser) Open(u string) error {
	ur, err := url.Parse(u)
	if err != nil {
		return err
	}
	return bow.httpGET(ur, nil)
}

// OpenForm appends the data values to the given URL and sends a GET request.
func (bow *Browser) OpenForm(u string, data url.Values) error {
	ul, err := url.Parse(u)
	if err != nil {
		return err
	}
	ul.RawQuery = data.Encode()

	return bow.Open(ul.String())
}

// OpenBookmark calls Open() with the URL for the bookmark with the given name.
func (bow *Browser) OpenBookmark(name string) error {
	url, err := bow.bookmarks.Read(name)
	if err != nil {
		return err
	}
	return bow.Open(url)
}

// Post requests the given URL using the POST method.
func (bow *Browser) Post(u string, contentType string, body io.Reader) error {
	ur, err := url.Parse(u)
	if err != nil {
		return err
	}
	return bow.httpPOST(ur, nil, contentType, body)
}

// PostForm requests the given URL using the POST method with the given data.
func (bow *Browser) PostForm(u string, data url.Values) error {
	return bow.Post(u, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

// Back loads the previously requested page.
//
// Returns a boolean value indicating whether a previous page existed, and was
// successfully loaded.
func (bow *Browser) Back() bool {
	if bow.history.Len() > 1 {
		bow.state = bow.history.Pop()
		bow.logDebug("Back called. New page is %s.", bow.state.Request.URL.String())
		return true
	}
	return false
}

// Reload duplicates the last successful request.
func (bow *Browser) Reload() error {
	if bow.state.Request != nil {
		bow.logDebug("Reloading page %s.", bow.state.Request.URL.String())
		return bow.httpRequest(bow.state.Request)
	}
	return errors.NewPageNotLoaded("Cannot reload, the previous request failed.")
}

// Recorder returns the set recorder.
func (bow *Browser) Recorder() jar.Recorder {
	return bow.recorder
}

// Bookmark saves the page URL in the bookmarks with the given name.
func (bow *Browser) Bookmark(name string) error {
	bow.logDebug("Bookmarking page %s.", bow.Url().String())
	return bow.bookmarks.Save(name, bow.ResolveUrl(bow.Url()).String())
}

// Click clicks on the page element matched by the given expression.
//
// Currently this is only useful for click on links, which will cause the browser
// to load the page pointed at by the link. Future versions of Surf may support
// JavaScript and clicking on elements will fire the click event.
func (bow *Browser) Click(expr string) error {
	sel := bow.Find(expr)
	if sel.Length() == 0 {
		return errors.NewElementNotFound(
			"Element not found matching expr '%s'.", expr)
	}
	if !sel.Is("a") {
		return errors.NewElementNotFound(
			"Expr '%s' must match an anchor tag.", expr)
	}
	href, err := bow.attrToResolvedUrl("href", sel)
	if err != nil {
		return err
	}
	bow.doClick(href)

	return bow.httpGET(href, bow.Url())
}

// Form returns the form in the current page that matches the given expr.
func (bow *Browser) Form(expr string) (Submittable, error) {
	sel := bow.Find(expr)
	if sel.Length() == 0 {
		return nil, errors.NewElementNotFound(
			"Form not found matching expr '%s'.", expr)
	}
	if !sel.Is("form") {
		return nil, errors.NewElementNotFound(
			"Expr '%s' does not match a form tag.", expr)
	}

	return bow.newForm(sel), nil
}

// Forms returns an array of every form in the page.
func (bow *Browser) Forms() []Submittable {
	sel := bow.Find("form")
	len := sel.Length()
	if len == 0 {
		return nil
	}

	forms := make([]Submittable, len)
	sel.Each(func(_ int, s *goquery.Selection) {
		forms = append(forms, bow.newForm(s))
	})
	return forms
}

// Links returns an array of every link found in the page.
func (bow *Browser) Links() []*Link {
	links := make([]*Link, 0, InitialAssetsSliceSize)
	bow.Find("a").Each(func(_ int, s *goquery.Selection) {
		href, err := bow.attrToResolvedUrl("href", s)
		if err == nil {
			links = append(links, NewLinkAsset(
				href,
				attrOrDefault("id", "", s),
				s.Text(),
			))
		}
	})

	return links
}

// Images returns an array of every image found in the page.
func (bow *Browser) Images() []*Image {
	images := make([]*Image, 0, InitialAssetsSliceSize)
	bow.Find("img").Each(func(_ int, s *goquery.Selection) {
		src, err := bow.attrToResolvedUrl("src", s)
		if err == nil {
			images = append(images, NewImageAsset(
				src,
				attrOrDefault("id", "", s),
				attrOrDefault("alt", "", s),
				attrOrDefault("title", "", s),
			))
		}
	})

	return images
}

// Stylesheets returns an array of every stylesheet linked to the document.
func (bow *Browser) Stylesheets() []*Stylesheet {
	stylesheets := make([]*Stylesheet, 0, InitialAssetsSliceSize)
	bow.Find("link").Each(func(_ int, s *goquery.Selection) {
		rel, ok := s.Attr("rel")
		if ok && rel == "stylesheet" {
			href, err := bow.attrToResolvedUrl("href", s)
			if err == nil {
				stylesheets = append(stylesheets, NewStylesheetAsset(
					href,
					attrOrDefault("id", "", s),
					attrOrDefault("media", "all", s),
					attrOrDefault("type", "text/css", s),
				))
			}
		}
	})

	return stylesheets
}

// Scripts returns an array of every script linked to the document.
func (bow *Browser) Scripts() []*Script {
	scripts := make([]*Script, 0, InitialAssetsSliceSize)
	bow.Find("script").Each(func(_ int, s *goquery.Selection) {
		src, err := bow.attrToResolvedUrl("src", s)
		if err == nil {
			scripts = append(scripts, NewScriptAsset(
				src,
				attrOrDefault("id", "", s),
				attrOrDefault("type", "text/javascript", s),
			))
		}
	})

	return scripts
}

// SiteCookies returns the cookies for the current site.
func (bow *Browser) SiteCookies() []*http.Cookie {
	return bow.cookies.Cookies(bow.Url())
}

// SetCookieJar is used to set the cookie jar the browser uses.
func (bow *Browser) SetCookieJar(cj http.CookieJar) {
	bow.cookies = cj
}

// SetUserAgent sets the user agent.
func (bow *Browser) SetUserAgent(userAgent string) {
	bow.userAgent = userAgent
}

// SetAttribute sets a browser instruction attribute.
func (bow *Browser) SetAttribute(a Attribute, v bool) {
	bow.attributes[a] = v
}

// SetAttributes is used to set all the browser attributes.
func (bow *Browser) SetAttributes(a AttributeMap) {
	bow.attributes = a
}

// SetAuthorization sets the username and password to use during authorization.
func (bow *Browser) SetAuthorization(username, password string) {
	bow.auth = Authorization{
		Username: username,
		Password: password,
	}
}

// SetBookmarksJar sets the bookmarks jar the browser uses.
func (bow *Browser) SetBookmarksJar(bj jar.Bookmarks) {
	bow.bookmarks = bj
}

// SetHistoryJar is used to set the history jar the browser uses.
func (bow *Browser) SetHistoryJar(hj jar.History) {
	bow.history = hj
}

// SetRecorderJar sets a jar.RecorderJar that will record browser states.
func (bow *Browser) SetRecorderJar(rj jar.Recorder) {
	bow.recorder = rj

	// Let the recorder know when the browser has made a request so it can
	// be recorded.
	bow.On(event.PostRequest, rj)

	// Have the recorder let the browser know when it's playing back requests
	// so the browser can make the request.
	bow.recorder.OnFunc(event.RecordReplay, (event.HandlerFunc)(func(_ event.Event, _, args interface{}) error {
		req := args.(*http.Request)
		bow.logDebug("Playing back request %s %s.", req.Method, req.URL.String())
		err := bow.httpRequest(req)
		if err != nil {
			return err
		}
		return nil
	}))
}

// SetHeadersJar sets the headers the browser sends with each request.
func (bow *Browser) SetHeadersJar(h http.Header) {
	bow.headers = h
}

// SetEventDispatcher sets the event dispatcher.
func (bow *Browser) SetEventDispatcher(ed event.Eventable) {
	bow.Dispatcher = ed.(*event.Dispatcher)
}

// SetLogger sets the instance that will be used to log events.
func (bow *Browser) SetLogger(l *log.Logger, lev LogLevel) {
	bow.logger = l
	bow.logLevel = lev
}

// AddRequestHeader sets a header the browser sends with each request.
func (bow *Browser) AddRequestHeader(name, value string) {
	bow.headers.Add(name, value)
}

// ResolveUrl returns an absolute URL for a possibly relative URL.
func (bow *Browser) ResolveUrl(u *url.URL) *url.URL {
	return bow.Url().ResolveReference(u)
}

// ResolveStringUrl works just like ResolveUrl, but the argument and return value are strings.
func (bow *Browser) ResolveStringUrl(u string) (string, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	pu = bow.Url().ResolveReference(pu)
	return pu.String(), nil
}

// Download writes the contents of the document to the given writer.
func (bow *Browser) Download(o io.Writer) (int64, error) {
	bow.logInfo("Downloading page %s.", bow.Url().String())
	h, err := bow.state.Dom.Html()
	if err != nil {
		return 0, err
	}
	l, err := io.WriteString(o, h)
	return int64(l), err
}

// Url returns the page URL as a string.
func (bow *Browser) Url() *url.URL {
	return bow.state.Request.URL
}

// StatusCode returns the response status code.
func (bow *Browser) StatusCode() int {
	return bow.state.Response.StatusCode
}

// Title returns the page title.
func (bow *Browser) Title() string {
	return bow.state.Dom.Find("title").Text()
}

// ResponseHeaders returns the page headers.
func (bow *Browser) ResponseHeaders() http.Header {
	return bow.state.Response.Header
}

// Body returns the page body as a string of html.
func (bow *Browser) Body() string {
	body, _ := bow.state.Dom.Find("body").Html()
	return body
}

// Dom returns the inner *goquery.Selection.
func (bow *Browser) Dom() *goquery.Selection {
	return bow.state.Dom.First()
}

// Find returns the dom selections matching the given expression.
func (bow *Browser) Find(expr string) *goquery.Selection {
	return bow.state.Dom.Find(expr)
}

// -- Unexported methods --

// buildClient creates, configures, and returns a *http.Client type.
func (bow *Browser) buildClient() *http.Client {
	client := &http.Client{}
	client.Jar = bow.cookies
	client.CheckRedirect = bow.shouldRedirect
	return client
}

// buildRequest creates and returns a *http.Request type.
// Sets any headers that need to be sent with the request.
func (bow *Browser) buildRequest(method string, u *url.URL, ref *url.URL) (*http.Request, error) {
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = bow.headers
	req.Header.Add("User-Agent", bow.userAgent)
	bow.logDebug("Setting User-Agent header to %s.", bow.userAgent)
	if bow.attributes[SendReferer] && ref != nil {
		bow.logDebug("Setting Referer header to %s.", ref.String())
		req.Header.Add("Referer", ref.String())
	}
	if bow.auth.Username != "" {
		auth := basicAuth(bow.auth.Username, bow.auth.Password)
		bow.logDebug("Setting Authorization header to %s.", auth)
		req.Header.Add(
			"Authorization",
			"Basic "+auth,
		)
	}

	return req, nil
}

// httpGET makes an HTTP GET request for the given URL.
// When via is not nil, and AttributeSendReferer is true, the Referer header will
// be set to ref.
func (bow *Browser) httpGET(u *url.URL, ref *url.URL) error {
	req, err := bow.buildRequest("GET", u, ref)
	if err != nil {
		return err
	}
	return bow.httpRequest(req)
}

// httpPOST makes an HTTP POST request for the given URL.
// When via is not nil, and AttributeSendReferer is true, the Referer header will
// be set to ref.
func (bow *Browser) httpPOST(u *url.URL, ref *url.URL, contentType string, body io.Reader) error {
	req, err := bow.buildRequest("POST", u, ref)
	if err != nil {
		return err
	}
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	req.Body = rc
	req.Header.Add("Content-Type", contentType)

	return bow.httpRequest(req)
}

// send uses the given *http.Request to make an HTTP request.
func (bow *Browser) httpRequest(req *http.Request) error {
	err := bow.doPreRequest(req)
	if err != nil {
		return bow.logError(err)
	}

	if bow.refresh != nil {
		bow.refresh.Stop()
	}
	bow.logInfo("Sending request. %s %s", req.Method, req.URL.String())
	resp, err := bow.buildClient().Do(req)
	if err != nil {
		return bow.logError(err)
	}
	bow.logInfo("Received %d response.", resp.StatusCode)

	dom, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return bow.logError(err)
	}
	bow.history.Push(bow.state)
	bow.state = jar.NewHistoryState(req, resp, dom)
	bow.handleMetaRefresh()

	err = bow.doPostRequest(resp)
	if err != nil {
		return bow.logError(err)
	}
	return nil
}

// handleMetaRefresh handles the meta refresh tag in the page.
func (bow *Browser) handleMetaRefresh() {
	if bow.attributes[MetaRefreshHandling] {
		sel := bow.Find("meta[http-equiv='refresh']")
		if sel.Length() > 0 {
			attr, ok := sel.Attr("content")
			if ok {
				dur, err := time.ParseDuration(attr + "s")
				if err == nil {
					bow.logDebug("Creating %d second timer for meta refresh.", dur.Seconds())
					bow.refresh = time.NewTimer(dur)
					go func() {
						<-bow.refresh.C
						bow.Reload()
					}()
				}
			}
		}
	}
}

// doPreRequest triggers the PreRequestEvent event.
func (bow *Browser) doPreRequest(req *http.Request) error {
	bow.logDebug("Doing event event.PreRequest.")
	return bow.Do(event.PreRequest, bow, req)
}

// doPostRequest triggers the PostRequestEvent event.
func (bow *Browser) doPostRequest(resp *http.Response) error {
	bow.logDebug("Doing event event.PostRequest.")
	return bow.Do(event.PostRequest, bow, resp)
}

// doClick triggers the ClickEvent event.
func (bow *Browser) doClick(u *url.URL) error {
	bow.logDebug("Doing event event.Click.")
	return bow.Do(event.Click, bow, u)
}

// newForm creates and returns a new *Form instance with the event.Submit event
// bound to the browser.
func (bow *Browser) newForm(s *goquery.Selection) *Form {
	form := NewForm(s)
	form.OnFunc(event.Submit, (event.HandlerFunc)(func(_ event.Event, sender, args interface{}) error {
		fm := sender.(*Form)
		action := bow.ResolveUrl(fm.Action())
		values := args.(url.Values)

		bow.logInfo("Submitting form %s %s %#v.", fm.Method(), fm.Action(), values)
		if fm.Method() == "GET" {
			return bow.OpenForm(action.String(), values)
		} else {
			return bow.PostForm(action.String(), values)
		}
	}))

	return form
}

// shouldRedirect is used as the value to http.Client.CheckRedirect.
func (bow *Browser) shouldRedirect(req *http.Request, _ []*http.Request) error {
	if bow.attributes[FollowRedirects] {
		return nil
	}
	return errors.NewLocation(
		"Redirects are disabled. Cannot follow '%s'.", req.URL.String())
}

// attributeToUrl reads an attribute from an element and returns a url.
func (bow *Browser) attrToResolvedUrl(name string, sel *goquery.Selection) (*url.URL, error) {
	src, ok := sel.Attr(name)
	if !ok {
		return nil, errors.NewAttributeNotFound(
			"Attribute '%s' not found.", name)
	}
	ur, err := url.Parse(src)
	if err != nil {
		return nil, err
	}

	return bow.ResolveUrl(ur), nil
}

// logDebug logs the given message using log.Printf at the debug level when a logger has been set.
//
// A new line is automatically appended to the message.
func (bow *Browser) logDebug(msg string, a ...interface{}) {
	if bow.logger != nil && bow.logLevel >= LogLevelDebug {
		bow.logger.Printf(msg+"\n", a...)
	}
}

// logInfo logs the given message using log.Printf at the info level when a logger has been set.
//
// A new line is automatically appended to the message.
func (bow *Browser) logInfo(msg string, a ...interface{}) {
	if bow.logger != nil && bow.logLevel >= LogLevelInfo {
		bow.logger.Printf(msg+"\n", a...)
	}
}

// logError logs the given message using log.Printf at the error level when a logger has been set.
//
// A new line is automatically appended to the message.
func (bow *Browser) logError(e error) error {
	if bow.logger != nil && bow.logLevel >= LogLevelError {
		bow.logger.Println(e.Error())
	}
	return e
}

// attributeOrDefault reads an attribute and returns it or the default value when it's empty.
func attrOrDefault(name, def string, sel *goquery.Selection) string {
	a, ok := sel.Attr(name)
	if ok {
		return a
	}
	return def
}

// basicAuth creates an Authentication header from the username and password.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
