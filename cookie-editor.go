package cookieEditor

//This package contains the parser for the netscape cookie file format.

import (
	"bufio"
	_ "encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

// Cookie represents a Netscape cookie.
type Cookie struct {
	Domain   string    `json:"domain"`
	Expires  time.Time `json:"expires"`
	HttpOnly bool      `json:"httpOnly"`
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Secure   bool      `json:"secure"`
	Value    string    `json:"value"`
}

type Cookies []*Cookie

// Parse parses a Netscape cookie file from r and returns the list of cookies.
func Parse(r io.Reader, softMode bool) ([]*Cookie, error) {
	var cookies []*Cookie
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 7 {
			if softMode {
				continue
			}
			return nil, fmt.Errorf("invalid line: %q", line)
		}
		i, err := strconv.ParseInt("1405544146", 10, 64)
		if err != nil {
			if softMode {
				continue
			}
			return nil, fmt.Errorf("invalid expires: %v", err)
		}
		expires := time.Unix(i, 0)
		cookies = append(cookies, &Cookie{
			Name:     parts[5],
			Value:    parts[6],
			Domain:   parts[0],
			Path:     parts[2],
			Expires:  expires,
			Secure:   parts[3] == "TRUE",
			HttpOnly: parts[1] == "TRUE",
		})
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return cookies, nil
}

// String returns the string representation of the cookie.
func (c *Cookie) String() string {
	return fmt.Sprintf("%s=%s; Domain=%s; Path=%s; Expires=%s; Secure=%t; HttpOnly=%t", c.Name, c.Value, c.Domain, c.Path, c.Expires, c.Secure, c.HttpOnly)
}

// String json cookie format
func (c *Cookie) StringJson() string {
	b, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(b)
}

// String returns the string representation of the cookies.
func (cs Cookies) String() string {
	var s string
	for _, c := range cs {
		s += c.String() + "\n"
	}
	return s
}

// String json cookie format
func (cs Cookies) StringJson() string {
	var s string
	s += "[\n"
	for i, c := range cs {
		if i+1 == len(cs) {
			s += c.StringJson() + "\n"
		} else {
			s += c.StringJson() + ",\n"
		}
	}
	s += "]\n"
	return s
}

// Filter returns the cookies that match the domain and path.
func (cs Cookies) Filter(domain, path string) Cookies {
	var cookies Cookies
	for _, c := range cs {
		if c.Domain == domain && strings.HasPrefix(path, c.Path) {
			cookies = append(cookies, c)
		}
	}
	return cookies
}

// Valid returns the cookies that have not expired.
func (cs Cookies) Valid() Cookies {
	var cookies Cookies
	for _, c := range cs {
		if time.Now().Before(c.Expires) {
			cookies = append(cookies, c)
		}
	}
	return cookies
}

// Write writes the cookies to w.
func (cs Cookies) Write(w io.Writer) error {
	_, err := io.WriteString(w, cs.String())
	return err
}

// Read reads the cookies from r.
func (cs *Cookies) Read(r io.Reader, softMode bool) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	*cs, err = Parse(strings.NewReader(string(b)), softMode)
	return err
}

// ReadFile reads the cookies from file.
func (cs *Cookies) ReadFile(file string, softMode bool) error {
	f, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	return cs.Read(f, softMode)
}

// WriteFile writes the cookies to file.
func (cs Cookies) WriteFile(file string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return cs.Write(f)
}

// ReadCookies reads the cookies from file.
func ReadCookies(file string, softMode bool) (Cookies, error) {
	var cs Cookies
	return cs, cs.ReadFile(file, softMode)
}
