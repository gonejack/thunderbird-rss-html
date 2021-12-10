package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"net/mail"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/kong"
	"github.com/gonejack/email"
	"github.com/yosssi/gohtml"
)

type options struct {
	About bool     `help:"About."`
	EML   []string `arg:"" optional:""`
}
type Converter struct {
	options
}

func (c *Converter) Run() (err error) {
	kong.Parse(&c.options,
		kong.Name("thunderbird-rss-html"),
		kong.Description("This command line converts thuderbird's exported RSS .eml file to .html file"),
		kong.UsageOnError(),
	)

	if c.About {
		fmt.Println("Visit https://github.com/gonejack/thunderbird-rss-html")
		return
	}

	// support Windows globbing
	if runtime.GOOS == "windows" {
		for _, eml := range c.EML {
			if eml == "*.eml" {
				c.EML = nil
				break
			}
		}
	}

	if len(c.EML) == 0 || c.EML[0] == "*.eml" {
		c.EML, _ = filepath.Glob("*.eml")
	}

	return c.run()
}
func (c *Converter) run() (err error) {
	if len(c.EML) == 0 {
		return errors.New("no .eml file given")
	}

	for _, eml := range c.EML {
		err = c.process(eml)
		if err != nil {
			err = fmt.Errorf("parse %s failed: %s", eml, err)
			return
		}
	}

	return
}
func (c *Converter) process(name string) (err error) {
	f, err := os.Open(name)
	if err != nil {
		return
	}
	defer f.Close()

	eml, err := email.NewEmailFromReader(f)
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(eml.HTML))
	if err != nil {
		return
	}
	c.patchHTML(eml, doc)

	htm, err := doc.Html()
	if err != nil {
		return
	}
	htm = gohtml.Format(htm)

	target := strings.TrimSuffix(name, ".eml") + ".html"
	return os.WriteFile(target, []byte(htm), 0766)
}
func (c *Converter) patchHTML(eml *email.Email, doc *goquery.Document) {
	c.patchHeader(eml, doc)
	c.patchFooter(eml, doc)
}
func (c *Converter) patchHeader(email *email.Email, doc *goquery.Document) {
	datetime := time.Now()
	pt, err := mail.ParseDate(email.Headers.Get("Date"))
	if err == nil {
		datetime = pt
	}

	meta := fmt.Sprintf(`<meta name="inostar:publish" content="%s">`, datetime.Format(time.RFC1123Z))
	doc.Find("head").AppendHtml(meta)

	const tpl = `
<p>
	<a title="Published: {published}" href="{link}" style="display:block; color: #000; padding-bottom: 10px; text-decoration: none; font-size:1em; font-weight: normal;">
		<span style="font-size: 1.5em;">{title}</span>
	</a>
</p>`

	replacer := strings.NewReplacer(
		"{link}", email.Headers.Get("Content-Base"),
		"{published}", datetime.Format("2006-01-02 15:04:05"),
		"{title}", html.EscapeString(email.Subject),
	)

	doc.Find("body").PrependHtml(replacer.Replace(tpl))
}
func (c *Converter) patchFooter(email *email.Email, doc *goquery.Document) {
	const tpl = `
<br/><br/>
<a style="display: block; display: inline-block; border-top: 1px solid #ccc; padding-top: 5px; color: #666; text-decoration: none;"
   href="{link}">{link}</a>
<p style="color:#999;">Save with <a style="color:#666; text-decoration:none; font-weight: bold;" 
									href="https://github.com/gonejack/thunderbird-rss-html">thunderbird-rss-html</a>
</p>`

	rpl := strings.NewReplacer(
		"{link}", email.Headers.Get("Content-Base"),
	)

	doc.Find("body").AppendHtml(rpl.Replace(tpl))
}
