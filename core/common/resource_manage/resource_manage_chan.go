package resource_manage

// ResourceManageChan inherits the ResourceManage interface.
// In spider, ResourceManageChan manage resource of Coroutine to crawl page.
// 基于 channel 的资源管理（利用 channel 的容量实现阻塞功能）
type ResourceManageChan struct {
    capnum uint
    mc     chan uint
}

// NewResourceManageChan returns initialized ResourceManageChan object which contains a resource pool.
// The num is the resource limit.
func NewResourceManageChan(num uint) *ResourceManageChan {
    mc := make(chan uint, num)
    return &ResourceManageChan{mc: mc, capnum: num}
}

// The GetOne apply for one resource.
// If resource pool is empty, current coroutine will be blocked.
func (this *ResourceManageChan) GetOne() {
    this.mc <- 1
}

// The FreeOne free resource and return it to resource pool.
func (this *ResourceManageChan) FreeOne() {
    <-this.mc
}

// The Has query for how many resource has been used.
func (this *ResourceManageChan) Used() uint {
    return uint(len(this.mc))
}

// The Left query for how many resource left in the pool.
func (this *ResourceManageChan) Left() uint {
    return this.capnum - uint(len(this.mc))
}
