package main

// Storage 存储接口
type Storage interface {

	// Shorten 将正常地址转成短地址
	Shorten(url string, exp int64) (string, error)

	// ShortlinkInfo 获取短地址信息
	ShortlinkInfo(eid string) (interface{}, error)

	// Unshorten 将短地址转换正常地址
	Unshorten(eid string) (string, error)
}
