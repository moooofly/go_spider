// Package resource_manage implements a resource management.
package resource_manage

// ResourceManager 用于抽象针对资源的管理
type ResourceManager interface {
    GetOne()     // 获取一个可用资源
    FreeOne()    // 释放一个可用资源
    Used() uint  // 当前被占用的资源数量
    Left() uint  // 当前剩余的资源数量
}
