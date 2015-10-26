package miniweb

import ()

type AnyFilter interface {
    Any(*Input, Output) bool
}

type AnyFunc func(*Input, Output) bool

func (f AnyFunc) Any(in *Input, out Output) bool {
    return f(in, out)
}

type OptionsFilter interface {
    Options(*Input, Output) bool
}

type OptionsFunc func(*Input, Output) bool

func (f OptionsFunc) Options(in *Input, out Output) bool {
    return f(in, out)
}

type HeadFilter interface {
    Head(*Input, Output) bool
}

type HeadFunc func(*Input, Output) bool

func (f HeadFunc) Head(in *Input, out Output) bool {
    return f(in, out)
}

type GetFilter interface {
    Get(*Input, Output) bool
}

type GetFunc func(*Input, Output) bool

func (f GetFunc) Get(in *Input, out Output) bool {
    return f(in, out)
}

type PostFilter interface {
    Post(*Input, Output) bool
}

type PostFunc func(*Input, Output) bool

func (f PostFunc) Post(in *Input, out Output) bool {
    return f(in, out)
}

type PutFilter interface {
    Put(*Input, Output) bool
}

type PutFunc func(*Input, Output) bool

func (f PutFunc) Put(in *Input, out Output) bool {
    return f(in, out)
}

type PatchFilter interface {
    Patch(*Input, Output) bool
}

type PatchFunc func(*Input, Output) bool

func (f PatchFunc) Patch(in *Input, out Output) bool {
    return f(in, out)
}

type DeleteFilter interface {
    Delete(*Input, Output) bool
}

type DeleteFunc func(*Input, Output) bool

func (f DeleteFunc) Delete(in *Input, out Output) bool {
    return f(in, out)
}

type TraceFilter interface {
    Trace(*Input, Output) bool
}

type TraceFunc func(*Input, Output) bool

func (f TraceFunc) Trace(in *Input, out Output) bool {
    return f(in, out)
}

type ConnectFilter interface {
    Connect(*Input, Output) bool
}

type ConnectFunc func(*Input, Output) bool

func (f ConnectFunc) Connect(in *Input, out Output) bool {
    return f(in, out)
}

type NotFoundFilter interface {
    NotFound(Output)
}

type NotFoundFunc func(Output)

func (f NotFoundFunc) NotFound(out Output) {
    f(out)
}