package pipeline

import (
    "github.com/moooofly/go_spider/core/common/com_interfaces"
    "github.com/moooofly/go_spider/core/common/page_items"
)

// 专门用于收集 PageItems 的 pipeline
type CollectPipelinePageItems struct {
    collector []*page_items.PageItems
}

func NewCollectPipelinePageItems() *CollectPipelinePageItems {
    collector := make([]*page_items.PageItems, 0)
    return &CollectPipelinePageItems{collector: collector}
}

// 添加收集内容
func (this *CollectPipelinePageItems) Process(items *page_items.PageItems, t com_interfaces.Task) {
    this.collector = append(this.collector, items)
}

// 获取当前收集结果信息
func (this *CollectPipelinePageItems) GetCollected() []*page_items.PageItems {
    return this.collector
}
