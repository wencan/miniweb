package miniweb

import (
    "fmt"
    "net/http"
    "regexp"
    "strings"
)

//与url相匹配的路由
type matchedRoute struct {
    Keys    []string      //Routing的Key
    Values  []string      //与Routing Key对应的路径片段
    Filters []interface{} //Routing的filters
}

//路由节点，或曰路由片段？
type routing struct {
    Key      string
    Filters  []interface{}
    Trailers []*routing
}

func (self *routing) match(path []string) (routes []matchedRoute, ok bool) {
    if len(path) == 0 {
        if len(self.Filters) > 0 {
            route := matchedRoute{Filters: self.Filters}
            routes = append(routes, route)
            ok = true
        }

        for _, trailer := range self.Trailers {
            if strings.HasPrefix(trailer.Key, "?:") && len(trailer.Filters) > 0 {
                route := matchedRoute{Filters: trailer.Filters}
                routes = append(routes, route)
                ok = true
            }
        }

        return
    }

    var segment string
    segment, path = path[0], path[1:]
    for _, trailer := range self.Trailers {
        if strings.HasPrefix(trailer.Key, "/") { //精确匹配
            key := strings.TrimPrefix(trailer.Key, "/")
            if key == segment {
                if rs, matched := trailer.match(path); matched {
                    routes = append(routes, rs...)
                    ok = true
                }
            }
        } else if trailer.Key == "*" { //*匹配
            if len(trailer.Trailers) > 0 {
                keys := []string{}
                values := []string{}
                var rs []matchedRoute

                pathCopy := append([]string{segment}, path...)
                for len(pathCopy) > 0 {
                    if rs, ok = trailer.match(pathCopy); ok {
                        for _, route := range rs {
                            route.Keys = append(keys, route.Keys...)
                            route.Values = append(values, route.Values...)
                            routes = append(routes, route)
                        }
                        break
                    }
                    segment, pathCopy = pathCopy[0], pathCopy[1:]
                    keys = append(keys, "")
                    values = append(values, segment)
                }
            } else {
                route := matchedRoute{Keys: []string{""}, Values: []string{segment}, Filters: trailer.Filters}
                for _, value := range path {
                    route.Keys = append(route.Keys, "")
                    route.Values = append(route.Values, value)
                }
                routes = append(routes, route)
                ok = true
            }
        } else if strings.HasPrefix(trailer.Key, ":") { //:匹配，含正则
            var key string

            if strings.Contains(trailer.Key, "(") && strings.HasSuffix(trailer.Key, ")") {
                reg := regexp.MustCompile("\\((?P<first>.+)\\)")
                subMatched := reg.FindStringSubmatch(trailer.Key)
                var pattern string
                if len(subMatched) >= 2 {
                    pattern = subMatched[1]
                }

                if len(pattern) == 0 {
                    break
                }
                if matched, _ := regexp.MatchString(pattern, segment); !matched {
                    break
                }

                reg = regexp.MustCompile("\\:(?P<second>.+)\\(")
                subMatched = reg.FindStringSubmatch(trailer.Key)
                if len(subMatched) >= 2 {
                    key = subMatched[1]
                }
            } else {
                key = strings.TrimPrefix(trailer.Key, ":")
            }

            if rs, matched := trailer.match(path); matched {
                for _, route := range rs {
                    route.Keys = append(route.Keys, key)
                    route.Values = append(route.Values, segment)
                    routes = append(routes, route)
                    ok = true
                }
            }
        } else if strings.HasPrefix(trailer.Key, "?:") { //?:匹配，含正则
            var key string
            var notMatched bool

            if strings.Contains(trailer.Key, "(") && strings.HasSuffix(trailer.Key, ")") {
                reg := regexp.MustCompile("\\((?P<first>.+)\\)")
                subMatched := reg.FindStringSubmatch(trailer.Key)
                var pattern string
                if len(subMatched) >= 2 {
                    pattern = subMatched[1]
                }
                if len(pattern) == 0 {
                    break
                }

                if matched, _ := regexp.MatchString(pattern, segment); matched {
                    reg = regexp.MustCompile("\\:(?P<second>.+)\\(")
                    subMatched := reg.FindStringSubmatch(trailer.Key)
                    if len(subMatched) >= 2 {
                        key = subMatched[1]
                    }
                } else {
                    notMatched = true
                }

            } else {
                key = strings.TrimPrefix(trailer.Key, "?:")
            }

            if !notMatched {
                if rs, matched := trailer.match(path); matched {
                    for _, route := range rs {
                        route.Keys = append(route.Keys, key)
                        route.Values = append(route.Values, segment)
                        routes = append(routes, route)
                        ok = true
                    }
                }
            }

            pathCopy := append([]string{segment}, path...)
            if rs, matched := trailer.match(pathCopy); matched {
                routes = append(routes, rs...)
                ok = true
            }

        } else if strings.HasPrefix(trailer.Key, "(") && strings.HasSuffix(trailer.Key, ")") { //正则匹配
            pattern := strings.TrimFunc(trailer.Key, func(c rune) bool {
                return c == '(' || c == ')'
            })

            if matched, _ := regexp.MatchString(pattern, segment); matched {
                if rs, matched := trailer.match(path); matched {
                    for _, route := range rs {
                        route.Keys = append([]string{""}, route.Keys...)
                        route.Values = append([]string{segment}, route.Values...)
                        routes = append(routes, route)
                    }
                    ok = true
                }
            }
        }
    }
    return
}

func (self *routing) addFilter(pattern []string, filter interface{}) {
    if len(pattern) == 0 {
        self.Filters = append(self.Filters, filter)
        return
    }

    var segment string
    segment, pattern = pattern[0], pattern[1:]

    //空目录也可匹配*
    //这里确保/和/*两个的filters的调用顺序
    if segment == "*" && len(pattern) == 0 {
        self.Filters = append(self.Filters, filter)
    }

    var node *routing
    //新加的必须在数组最后，以确保filters的调用顺序
    if len(self.Trailers) > 0 {
        last := self.Trailers[len(self.Trailers)-1]
        if last.Key == segment {
            node = last
        }
    }
    if node == nil {
        node = &routing{Key: segment}
        self.Trailers = append(self.Trailers, node)
    }

    node.addFilter(pattern, filter)
}

//路由器
type Router struct {
    root *routing //路由树根
}

func NewRouter() *Router {
    return &Router{&routing{}}
}

func (Router) filter_apply(filters []interface{}, in *Input, out Output) (ok bool, over bool) {
    for _, filter := range filters {
        if f, y := filter.(AnyFilter); y {
            over = f.Any(in, out)
            ok = true
            return
        }

        switch in.Request.Method {
        case "GET":
            if f, y := filter.(GetFilter); y {
                over = f.Get(in, out)
                ok = true
            }
        }
    }
    return
}

func (self Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    out := Output{w}

    var parts []string
    parts = strings.FieldsFunc(r.RequestURI, func(c rune) bool {
        return c == '?' || c == '#'
    })

    //解析url
    pathPart := parts[0]
    pathPart = strings.ToLower(pathPart)
    var path []string
    for _, segment := range strings.Split(pathPart, "/") {
        if len(segment) == 0 {
            continue
        }
        path = append(path, segment)
    }

    //解析查询参数
    querystrings := map[string][]string{}
    parts = strings.Split(r.RequestURI, "?")
    if len(parts) > 1 {
        trailer := parts[1]
        parts = strings.Split(trailer, "#")
        str := parts[0]

        parts = strings.FieldsFunc(str, func(c rune) bool {
            return c == '&' || c == ';'
        })
        for _, part := range parts {
            if len(part) == 0 {
                continue
            }
            kv := strings.Split(part, "=")
            key := kv[0]
            var value string
            if len(kv) > 1 {
                value = kv[1]
            }

            values, ok := querystrings[key]
            if ok {
                querystrings[key] = append(values, value)
            } else {
                querystrings[key] = []string{value}
            }
        }
    }

    //net/http的RequestURL不带fragment
    //    //解析片段
    //    var fragment string
    //    parts = strings.Split(r.RequestURI, "#")
    //    if len(parts) > 1 {
    //        fragment = parts[1]
    //    }

    //分派处理
    if routes, ok := self.root.match(path); ok {
        for _, route := range routes {
            fields := map[string]string{}
            count := 0
            for idx, key := range route.Keys {
                if len(key) == 0 {
                    count++
                    key = fmt.Sprintf("_%d", count)
                }
                fields[key] = route.Values[idx]
            }
            in := Input{Request: r, Fields: fields, QueryStrings: querystrings /*, Fragment: fragment*/}

            if ok, over := self.filter_apply(route.Filters, &in, out); ok {
                if over {
                    return
                }
            }
        }
    } 
    
    self.NotFound(out)
}

//返回404
func (Router) NotFound(out Output) {
    out.Return(http.StatusNotFound, []byte(http.StatusText(http.StatusNotFound)))
}

//返回405
func (Router) MethodNotAllowed(out Output) {
    out.Return(http.StatusMethodNotAllowed, []byte(http.StatusText(http.StatusMethodNotAllowed)))
}

//添加一个filter
func (self Router) Filter(pattern string, filter interface{}) {
    pattern = strings.ToLower(pattern)

    var parts []string
    for _, part := range strings.Split(pattern, "/") {
        if len(part) == 0 {
            continue
        }
        switch part[0] {
        case ':':
        case '?':
        case '*':
        case '(':
        default:
            part = fmt.Sprintf("/%s", part)
        }
        parts = append(parts, part)
    }
    self.root.addFilter(parts, filter)
}

func (self Router) Any(pattern string, filter AnyFilter) {
    self.Filter(pattern, filter)
}

func (self Router) AnyFunc(pattern string, filter func(*Input, Output) bool) {
    self.Filter(pattern, AnyFunc(filter))
}

func (self Router) Options(pattern string, filter OptionsFilter) {
    self.Filter(pattern, filter)
}

func (self Router) OptionsFunc(pattern string, filter func(*Input, Output) bool) {
    self.Filter(pattern, OptionsFunc(filter))
}

func (self Router) Head(pattern string, filter HeadFilter) {
    self.Filter(pattern, filter)
}

func (self Router) HeadFunc(pattern string, filter func(*Input, Output) bool) {
    self.Filter(pattern, HeadFunc(filter))
}

func (self Router) Get(pattern string, filter GetFilter) {
    self.Filter(pattern, filter)
}

func (self Router) GetFunc(pattern string, filter func(*Input, Output) bool) {
    self.Filter(pattern, GetFunc(filter))
}

func (self Router) Post(pattern string, filter PostFilter) {
    self.Filter(pattern, filter)
}

func (self Router) PostFunc(pattern string, filter func(*Input, Output) bool) {
    self.Filter(pattern, PostFunc(filter))
}

func (self Router) Put(pattern string, filter PutFilter) {
    self.Filter(pattern, filter)
}

func (self Router) PutFunc(pattern string, filter func(*Input, Output) bool) {
    self.Filter(pattern, PutFunc(filter))
}

func (self Router) Patch(pattern string, filter PatchFilter) {
    self.Filter(pattern, filter)
}

func (self Router) PatchtFunc(pattern string, filter func(*Input, Output) bool) {
    self.Filter(pattern, PatchFunc(filter))
}

func (self Router) Delete(pattern string, filter DeleteFilter) {
    self.Filter(pattern, filter)
}

func (self Router) DeleteFunc(pattern string, filter func(*Input, Output) bool) {
    self.Filter(pattern, DeleteFunc(filter))
}

func (self Router) Trace(pattern string, filter TraceFilter) {
    self.Filter(pattern, filter)
}

func (self Router) TraceFunc(pattern string, filter func(*Input, Output) bool) {
    self.Filter(pattern, TraceFunc(filter))
}

func (self Router) Connect(pattern string, filter ConnectFilter) {
    self.Filter(pattern, filter)
}

func (self Router) ConnectFunc(pattern string, filter func(*Input, Output) bool) {
    self.Filter(pattern, ConnectFunc(filter))
}
