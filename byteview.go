package cache

//缓存视图,避免直接暴露byte数组而造成缓存修改
type ByteView struct {
	b []byte
}

//视图只读不涉及修改，因此结构体值绑定函数

//返回视图大小
func (v ByteView) Len() int {
	return len(v.b)
}

/*
外部提供的[]byte可以被修改，而视图要保证不可被修改,因此视图创建要clone
返回给外部的[]byte也要clone,保证外部修改不会影响视图
*/

func NewByteView(b []byte) ByteView {
	return ByteView{b: clone(b)}
}
//返回视图的byte切片副本
func (v ByteView) ByteSlice() []byte {
	return clone(v.b)
}

func clone(b []byte) []byte {
	c:=make([]byte,len(b))
	copy(c,b)
	return c
}
