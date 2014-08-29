// Package surf ensembles other packages into a usable browser.
package surf

import (
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	"github.com/headzoo/surf/event"
	"github.com/headzoo/surf/jar"
)

var (
	// DefaultUserAgent is the global user agent value.
	DefaultUserAgent = agent.Create()

	// DefaultSendRefererAttribute is the global value for the AttributeSendReferer attribute.
	DefaultSendReferer = true

	// DefaultMetaRefreshHandlingAttribute is the global value for the AttributeHandleRefresh attribute.
	DefaultMetaRefreshHandling = true

	// DefaultFollowRedirectsAttribute is the global value for the AttributeFollowRedirects attribute.
	DefaultFollowRedirects = true
)

// NewBrowser creates and returns a *browser.Browser type.
func NewBrowser() *browser.Browser {
	bow := &browser.Browser{}
	bow.SetEventDispatcher(event.NewDispatcher())
	bow.SetUserAgent(DefaultUserAgent)
	bow.SetCookieJar(jar.NewMemoryCookies())
	bow.SetBookmarksJar(jar.NewMemoryBookmarks())
	bow.SetHistoryJar(jar.NewMemoryHistory())
	bow.SetRecorderJar(jar.NewMemoryRecorder())
	bow.SetHeadersJar(jar.NewMemoryHeaders())
	bow.SetAttributes(browser.AttributeMap{
		browser.SendReferer:         DefaultSendReferer,
		browser.MetaRefreshHandling: DefaultMetaRefreshHandling,
		browser.FollowRedirects:     DefaultFollowRedirects,
	})

	return bow
}
